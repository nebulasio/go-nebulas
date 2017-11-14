// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//
'use strict';

const esprima = require('esprima.js');

function traverse(object, visitor, master) {
    var key, child, parent, path;

    parent = (typeof master === 'undefined') ? [] : master;

    if (visitor.call(null, object, parent) === false) {
        return;
    }

    for (key in object) {
        if (object.hasOwnProperty(key)) {
            child = object[key];
            if (Array.isArray(object)) {
                path = [];
            } else {
                path = [{
                    node: object,
                    key: key
                }];
            }
            path.push.apply(path, parent);
            if (typeof child === 'object' && child !== null) {
                traverse(child, visitor, path);
            }
        }
    }
};

const CodeGenerator = {
    InstructionIncr: function (context, item) {
        context.traceable_source += ";_instruction_counter.incr(" + item.value + ");";
    },
    NewStartBlockWithIncrStatement: function (context, item) {
        context.traceable_source += "{_instruction_counter.incr(" + item.value + ");"
    },
    NewEndBlockStatement: function (context, item) {
        context.traceable_source += ";}";
    }
};

const ExpectedSyntaxForTracing = {
    CallExpression: 1,
    AssignmentExpression: 1,
    BinaryExpression: 1,
    UpdateExpression: 1,
    UnaryExpression: 1,
    LogicalExpression: 1,
    MemberExpression: 1,
    NewExpression: 1,
    ThrowStatement: 1,
    MetaProperty: 1,
    ConditionalExpression: 1,
    YieldExpression: 1
};

const ExpectedAncenstorSyntaxForTracingProcessors = {
    ExpressionStatement: function (tracer, context, stmt, parent_stmt) {
        // check whether in BlockStatement.
        if (!parent_stmt || parent_stmt.node.type == 'BlockStatement' || parent_stmt.node.type == 'Program' || parent_stmt.node.type == 'SwitchCase' || parent_stmt.node.type == 'LabeledStatement') {
            tracer.record_generator(context, stmt.node.range[0], CodeGenerator.InstructionIncr);
        } else {
            // generate BlockStatement.
            tracer.record_generator(context, stmt.node.range[0], CodeGenerator.NewStartBlockWithIncrStatement);
            tracer.record_generator(context, stmt.node.range[1], CodeGenerator.NewEndBlockStatement);
        }
    },
    BlockStatement: function (tracer, context, stmt) {
        tracer.record_generator(context, stmt.node.range[0] + 1, CodeGenerator.InstructionIncr);
    },
    IfStatement: function (tracer, context, stmt, parent_stmt) {
        var target_nodes = [];

        if (stmt.key === 'test') {
            // test statement.
            target_nodes.push(stmt.node.consequent);
            target_nodes.push(stmt.node.alternate);
        } else if (stmt.key === 'consequent') {
            // consequent statement.
            target_nodes.push(stmt.node.consequent);
        } else if (stmt.key === 'alternate') {
            // alternate statement.
            target_nodes.push(stmt.node.consequent);
        } else {
            return;
        }

        for (var i = 0; i < target_nodes.length; i++) {
            var target_node = target_nodes[i];
            if (!target_node) {
                continue
            };
            if (target_node.type == 'BlockStatement') {
                return ExpectedAncenstorSyntaxForTracingProcessors.BlockStatement(tracer, context, {
                    node: target_node,
                    key: stmt.key
                });
            } else {
                // generate BlockStatement.
                tracer.record_generator(context, target_node.range[0], CodeGenerator.NewStartBlockWithIncrStatement);
                tracer.record_generator(context, target_node.range[1], CodeGenerator.NewEndBlockStatement);
            }
        }
    },
    SwitchStatement: function (tracer, context, stmt, parent_stmt) {
        var target_nodes = [];

        if (stmt.key === 'discriminant') {
            // discriminant statement.
            target_nodes.push.apply(target_nodes, stmt.node.cases);
        } else {
            return;
        }

        for (var i = 0; i < target_nodes.length; i++) {
            var consequent = target_nodes[i].consequent;
            if (consequent.length == 0) {
                continue;
            }

            tracer.record_generator(context, consequent[0].range[0], CodeGenerator.InstructionIncr);
        }
    },
    SwitchCase: function (tracer, context, stmt, parent_stmt) {
        var consequent = stmt.node.consequent;
        if (consequent.length == 0) {
            return;
        }

        tracer.record_generator(context, consequent[0].range[0], CodeGenerator.InstructionIncr);
    },
    ArrowFunctionExpression: function (tracer, context, stmt, parent_stmt) {
        // let foo = (s) => s+1
        // the body can't be BlockStatement.
        var target_node = stmt.node.body;
        // generate BlockStatement.
        tracer.record_generator(context, target_node.range[0], CodeGenerator.NewStartBlockWithIncrStatement);
        tracer.record_generator(context, target_node.range[1], CodeGenerator.NewEndBlockStatement);
    },
    DoWhileStatement: function (tracer, context, stmt, parent_stmt) {
        debugger
        var target_node = stmt.node.body;
        if (target_node.type == 'BlockStatement') {
            return ExpectedAncenstorSyntaxForTracingProcessors.BlockStatement(tracer, context, {
                node: target_node,
                key: stmt.key
            });
        } else {
            // generate BlockStatement.
            tracer.record_generator(context, target_node.range[0], CodeGenerator.NewStartBlockWithIncrStatement);
            tracer.record_generator(context, target_node.range[1], CodeGenerator.NewEndBlockStatement);
        }
    },
    ForStatement: function (tracer, context, stmt, parent_stmt) {
        if (stmt.key == 'init') {
            // put Incr at the beginning.
            return ExpectedAncenstorSyntaxForTracingProcessors.ExpressionStatement(tracer, context, stmt, parent_stmt);
        }
        // put Incr to the body.
        var target_node = stmt.node.body;
        // generate BlockStatement.
        tracer.record_generator(context, target_node.range[0], CodeGenerator.NewStartBlockWithIncrStatement);
        tracer.record_generator(context, target_node.range[1], CodeGenerator.NewEndBlockStatement);
    },
    ForInStatement: function (tracer, context, stmt, parent_stmt) {
        if (stmt.key == 'right') {
            // put Incr at the beginning.
            return ExpectedAncenstorSyntaxForTracingProcessors.ExpressionStatement(tracer, context, stmt, parent_stmt);
        } else if (stmt.key == 'left') {
            // ignore.
            return;
        }

        // put Incr to the body.
        var target_node = stmt.node.body;
        // generate BlockStatement.
        tracer.record_generator(context, target_node.range[0], CodeGenerator.NewStartBlockWithIncrStatement);
        tracer.record_generator(context, target_node.range[1], CodeGenerator.NewEndBlockStatement);
    },
    ForOfStatement: function (tracer, context, stmt, parent_stmt) {
        if (stmt.key == 'right') {
            // put Incr at the beginning.
            return ExpectedAncenstorSyntaxForTracingProcessors.ExpressionStatement(tracer, context, stmt, parent_stmt);
        } else if (stmt.key == 'left') {
            // ignore.
            return;
        }

        // put Incr to the body.
        var target_node = stmt.node.body;
        // generate BlockStatement.
        tracer.record_generator(context, target_node.range[0], CodeGenerator.NewStartBlockWithIncrStatement);
        tracer.record_generator(context, target_node.range[1], CodeGenerator.NewEndBlockStatement);
    },
    WhileStatement: function (tracer, context, stmt, parent_stmt) {
        if (stmt.key == 'test') {
            // put Incr at the beginning.
            return ExpectedAncenstorSyntaxForTracingProcessors.ExpressionStatement(tracer, context, stmt, parent_stmt);
        }

        // put Incr to the body.
        var target_node = stmt.node.body;
        // generate BlockStatement.
        tracer.record_generator(context, target_node.range[0], CodeGenerator.NewStartBlockWithIncrStatement);
        tracer.record_generator(context, target_node.range[1], CodeGenerator.NewEndBlockStatement);
    },
    WithStatement: function (tracer, context, stmt, parent_stmt) {
        if (stmt.key == 'object') {
            // put Incr at the beginning.
            return ExpectedAncenstorSyntaxForTracingProcessors.ExpressionStatement(tracer, context, stmt, parent_stmt);
        }

        // put Incr to the body.
        var target_node = stmt.node.body;
        // generate BlockStatement.
        tracer.record_generator(context, target_node.range[0], CodeGenerator.NewStartBlockWithIncrStatement);
        tracer.record_generator(context, target_node.range[1], CodeGenerator.NewEndBlockStatement);
    }
};


function Tracker(source, options) {
    this.source = source;
    this.records = new Map();
    this.options = options || {};
};

Tracker.prototype = {
    parse: function () {
        this.ast = esprima.parseScript(this.source, {
            range: true,
            log: true
        });
    },
    traverse: function () {
        traverse(this.ast, this.traverse_delegate.bind(this));
    },
    traverse_delegate: function (node, path) {
        if ((node.type in ExpectedSyntaxForTracing)) {
            var value = ExpectedSyntaxForTracing[node.type];
            this.process_node({
                node: node,
                path: path,
                value: value
            });
        }
    },
    process_node: function (context) {
        for (var i = 0; i < context.path.length; i++) {
            var ancestor = context.path[i];
            var processor = ExpectedAncenstorSyntaxForTracingProcessors[ancestor.node.type];
            if (processor) {
                return processor(this, context, ancestor, i + 1 == context.path.length ? null : context.path[i + 1]);
            }
        }
    },
    record_generator: function (context, position, generator) {
        var item = this.records.get(position);
        if (item === undefined || item == null) {
            item = {
                value: 0,
                position: position,
                generator: generator
            };
            this.records.set(position, item);
        }
        item.value += context.value;
    },
    generate: function () {
        var ordered_items = Array.from(this.records.values());
        ordered_items.sort(function (a, b) {
            return a.position - b.position;
        });

        var context = {
            start_offset: 0,
            traceable_source: "",
            source: this.source
        };
        ordered_items.forEach(function (item) {
            context.traceable_source += context.source.slice(context.start_offset, item.position);
            item.generator(context, item);
            context.start_offset = item.position;
        });
        context.traceable_source += context.source.slice(context.start_offset);
        this.traceable_source = context.traceable_source;

    },
    process: function () {
        this.parse();
        this.traverse();
        this.generate();
        return this.traceable_source;
    }
};

function processScript(source, options) {
    var tracer = new Tracker(source, options);
    return tracer.process();
};

exports["parseScript"] = esprima.parseScript;
exports["traverse"] = traverse;
exports["processScript"] = processScript;

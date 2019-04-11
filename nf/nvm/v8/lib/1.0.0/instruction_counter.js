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

const module_path_prefix = (typeof process !== 'undefined') && (process.release.name === 'node') ? './' : '';
const esprima = require(module_path_prefix + 'esprima.js');

function traverse(object, visitor, master, injection_context_from_parent) {
    var key, child, parent, path;

    parent = (typeof master === 'undefined') ? [{
        node: null,
        key: ""
    }] : master;

    var injection_context = visitor.call(null, object, parent, injection_context_from_parent);
    if (injection_context === false) {
        return;
    }

    for (key in object) {
        if (object.hasOwnProperty(key)) {
            child = object[key];
            if (typeof child === 'object' && child !== null) {
                var injection_context_of_key = injection_context ? injection_context[key] : injection_context_from_parent;

                if (Array.isArray(object)) {
                    // ignore Array object in parents.
                    path = [];
                } else {
                    path = [{
                        node: object,
                        key: key
                    }];
                }
                path.push.apply(path, parent);
                traverse(child, visitor, path, injection_context_of_key);
            }
        }
    }
};

// key is the Expression, value is the count of instruction of the Expression.
const TrackingExpressions = {
    CallExpression: 8,
    AssignmentExpression: 3,
    BinaryExpression: 3,
    UpdateExpression: 3,
    UnaryExpression: 3,
    LogicalExpression: 3,
    MemberExpression: 4,
    NewExpression: 8,
    ThrowStatement: 6,
    MetaProperty: 4,
    ConditionalExpression: 3,
    YieldExpression: 6,
};

const InjectableExpressions = {
    ExpressionStatement: 1,
    VariableDeclaration: 1,
    ReturnStatement: 1,
    ThrowStatement: 1,
};

const InjectionType = {
    BEFORE_NODE: "BEFORE_NODE",
    AT_BEGINNING: "AT_BEGINNING",
    INNER_BEGINNING: "INNER_BEGINNING",
    INNER_BEGINNING_NOT_AND_OR: "INNER_BEGINNING_NOT_AND_OR",
};

const InjectionCodeGenerators = {
    CounterIncrFunc: function (value) {
        return "_instruction_counter.incr(" + value + ");";
    },
    BlockStatementBeginAndCounterIncrFunc: function (value) {
        if (value > 0) {
            return "{_instruction_counter.incr(" + value + ");"
        } else {
            return "{";
        }
    },
    BlockStatementEndAndCounterIncrFunc: function (value) {
        if (value > 0) {
            return "_instruction_counter.incr(" + value + ");}"
        } else {
            return "}";
        }
    },
    BlockStatementBeginAndCounterIncrFuncAndReturn: function (value) {
        if (value > 0) {
            return "{_instruction_counter.incr(" + value + "); return "
        } else {
            return "{return ";
        }
    },
    BeginInnerCounterIncrFunc: function (value) {
        return "_instruction_counter.incr(" + value + ") && (";
    },
    EndInnerCounterIncrFunc: function (value) {
        return ")";
    },
    CounterIncrFuncUsingNotAndLogicalOrFunc: function (value) {
        return "!_instruction_counter.incr(" + value + ") || ";
    },
};

function InjectionContext(node, type) {
    this.node = node;
    this.type = type;
};

function record_injection_info(injection_records, pos, value, injection_func) {
    var item = injection_records.get(pos);
    if (!item) {
        item = {
            pos: pos,
            value: 0,
            func: injection_func,
        };
        injection_records.set(pos, item);
    }
    item.value += value;
};

function processScript(source, strictDisallowUsage) {
    var injection_records = new Map();
    var record_injection = function (pos, value, injection_func) {
        return record_injection_info(injection_records, pos, value, injection_func);
    };

    function ensure_block_statement(node) {
        if (!node || !node.type) {
            // not a valid node, ignore
            return;
        }

        if (!(node.type in {
                BlockStatement: "",
                IfStatement: "",
            })) {
            record_injection(node.range[0], 0, InjectionCodeGenerators.BlockStatementBeginAndCounterIncrFunc);
            record_injection(node.range[1], 0, InjectionCodeGenerators.BlockStatementEndAndCounterIncrFunc);
        }
    };

    var ast = esprima.parseScript(source, {
        range: true,
        loc: true
    });

    var source_line_offset = 0;

    traverse(ast, function (node, parents, injection_context_from_parent) {
        // throw error when "_instruction_counter" was redefined in source.
        disallowRedefineOfInstructionCounter(node, parents, strictDisallowUsage);

        // 1. flag find the injection point, eg a Expression/Statement can inject code directly.
        if (node.type == "IfStatement") {
            ensure_block_statement(node.consequent);
            ensure_block_statement(node.alternate);
            return {
                "test": new InjectionContext(node.test, InjectionType.INNER_BEGINNING),
            };
        } else if (node.type == "ForStatement") {
            debugger
            ensure_block_statement(node.body);
            return {
                "init": new InjectionContext(node, InjectionType.BEFORE_NODE),
                "test": new InjectionContext(node.test, InjectionType.INNER_BEGINNING),
                "update": new InjectionContext(node.update, InjectionType.INNER_BEGINNING),
            };
        } else if (node.type == "ForInStatement") {
            ensure_block_statement(node.body);

            // because for in just call right once and iterate internal,
            // to keep inst const consistency with others, we manually add 1.
            var body = node.body;
            var pos = body.range[0];
            if (body.type === 'BlockStatement') {
                pos = body.range[0] + 1;
            }
            record_injection(pos, 1, InjectionCodeGenerators.CounterIncrFunc);

            return {
                "left": new InjectionContext(node, InjectionType.BEFORE_NODE),
                "right": new InjectionContext(node, InjectionType.BEFORE_NODE),
            };
        } else if (node.type == "ForOfStatement") {
            ensure_block_statement(node.body);

            // because for in just call right once and iterate internal,
            // to keep inst const consistency with others, we manually add 1.
            var body = node.body;
            var pos = body.range[0];
            if (body.type === 'BlockStatement') {
                pos = body.range[0] + 1;
            }
            record_injection(pos, 1, InjectionCodeGenerators.CounterIncrFunc);

            return {
                "left": new InjectionContext(node, InjectionType.BEFORE_NODE),
                "right": new InjectionContext(node, InjectionType.BEFORE_NODE),
            };
        } else if (node.type == "WhileStatement") {
            ensure_block_statement(node.body);
            return {
                "test": new InjectionContext(node.test, InjectionType.INNER_BEGINNING),
            };
        } else if (node.type == "DoWhileStatement") {
            ensure_block_statement(node.body);
            return {
                "test": new InjectionContext(node.test, InjectionType.INNER_BEGINNING),
            };
        } else if (node.type == "WithStatement") {
            ensure_block_statement(node.body);
            return {
                "object": new InjectionContext(node, InjectionType.BEFORE_NODE),
            };
        } else if (node.type == "SwitchStatement") {
            return {
                "discriminant": new InjectionContext(node, InjectionType.BEFORE_NODE),
            };
        } else if (node.type == "ArrowFunctionExpression") {
            var body = node.body;
            if (body.type !== 'BlockStatement') {
                record_injection(body.range[0], 0, InjectionCodeGenerators.BlockStatementBeginAndCounterIncrFuncAndReturn);
                record_injection(body.range[1], 0, InjectionCodeGenerators.BlockStatementEndAndCounterIncrFunc);

                // only return injection context when body is not in {};
                return {
                    "body": new InjectionContext(body, InjectionType.BEFORE_NODE),
                };
            }
        } else if (node.type == "ConditionalExpression") {
            return {
                "test": new InjectionContext(node.test, InjectionType.INNER_BEGINNING_NOT_AND_OR),
                "consequent": new InjectionContext(node.consequent, InjectionType.INNER_BEGINNING_NOT_AND_OR),
                "alternate": new InjectionContext(node.alternate, InjectionType.INNER_BEGINNING_NOT_AND_OR),
            };
        } else {

            // Other Expressions.
            var tracing_val = TrackingExpressions[node.type];
            if (!tracing_val) {
                // not the tracking expression, ignore.
                return;
            }

            // If no parent, apply default rule: BEFORE_NODE.
            var parent_node = parents[0].node;
            if (!parent_node) {
                record_injection(node.range[0], tracing_val, InjectionCodeGenerators.CounterIncrFunc);
                return;
            }

            var injection_type = null;
            var target_node = null;

            if (injection_context_from_parent) {
                target_node = injection_context_from_parent.node;
                injection_type = injection_context_from_parent.type;
            } else {
                injection_type = InjectionType.BEFORE_NODE;
            }

            if (!target_node) {
                if (node.type in InjectableExpressions) {
                    target_node = node;
                } else {
                    // searching parent to find the injection position.
                    for (var i = 0; i < parents.length; i++) {
                        var ancestor = parents[i];
                        if (ancestor.node.type in InjectableExpressions) {
                            target_node = ancestor.node;
                            break;
                        }
                    }
                }
            }

            var pos = -1;
            var generator = InjectionCodeGenerators.CounterIncrFunc;

            switch (injection_type) {
                case InjectionType.BEFORE_NODE:
                    pos = target_node.range[0];
                    break;
                case InjectionType.AT_BEGINNING:
                    if (target_node.type === 'BlockStatement') {
                        pos = target_node.range[0] + 1; // after "{".
                    } else {
                        pos = target_node.range[0]; // before statement start.
                    }
                    break;
                case InjectionType.INNER_BEGINNING:
                    pos = -1;
                    record_injection(target_node.range[0], tracing_val, InjectionCodeGenerators.BeginInnerCounterIncrFunc);
                    record_injection(target_node.range[1], tracing_val, InjectionCodeGenerators.EndInnerCounterIncrFunc);
                    break;
                case InjectionType.INNER_BEGINNING_NOT_AND_OR:
                    pos = target_node.range[0];
                    generator = InjectionCodeGenerators.CounterIncrFuncUsingNotAndLogicalOrFunc;
                    break;
                default:
                    throw new Error("Unknown Injection Type " + injection_type);
            }

            if (pos >= 0) {
                record_injection(pos, tracing_val, generator);
            }
        }
    });

    // generate traceable source.
    var ordered_records = Array.from(injection_records.values());
    ordered_records.sort(function (a, b) {
        return a.pos - b.pos;
    });


    var start_offset = 0,
        traceable_source = "";
    ordered_records.forEach(function (record) {
        traceable_source += source.slice(start_offset, record.pos);
        traceable_source += record.func(record.value);
        start_offset = record.pos;
    });
    traceable_source += source.slice(start_offset);

    return {
        traceableSource: traceable_source,
        lineOffset: source_line_offset
    };
};

// throw error when "_instruction_counter" was redefined.
function disallowRedefineOfInstructionCounter(node, parents, strictDisallowUsage) {
    if (node.type == 'Identifier') {
        if (node.name != '_instruction_counter') {
            return;
        }
    } else if (node.type == 'Literal') {
        if (node.value != '_instruction_counter') {
            return;
        }
    } else {
        return;
    }

    if (strictDisallowUsage) {
        throw new Error("redefine or use _instruction_counter are now allowed.");
    }

    var parent_node = parents[0].node;
    if (!parent_node) {
        return;
    }

    if (parent_node.type in {
            VariableDeclarator: "",
            FunctionDeclaration: "",
            FunctionExpression: "",
            ArrayPattern: "",
        }) {
        throw new Error("redefine _instruction_counter is now allowed.");
    }
};


exports["parseScript"] = esprima.parseScript;
exports["processScript"] = processScript;


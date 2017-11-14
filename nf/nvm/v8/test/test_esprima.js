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

var esprima = require('../lib/esprima.js');

var source = "'use strict';\nvar SampleContract = function() {\n    LocalContractStorage.defineProperties(this, {\n        name: null,\n        count: null\n    });\n    LocalContractStorage.defineMapProperty(this, \"allocation\" + 123);\n    this.a = 0;\n};";
var x = "\nSampleContract.prototype = {\n    init: function(name, count, allocation) {\n        this.name = name;\n        this.count = count;\n        allocation.forEach(function(item) {\n            this.allocation.put(item.name, item.count);\n        }, this);\n    },\n    dump: function() {\n        console.log('dump: this.name = ' + this.name);\n        console.log('dump: this.count = ' + this.count);\n        return this.a;\n    },\n    incr: function() {\n        this.a++;\n        var z = this.a;\n        console.log(this.dump());\n        return this.a;\n    },\n    verify: function(expectedName, expectedCount, expectedAllocation) {\n        if (!Object.is(this.name, expectedName)) {\n            throw new Error(\"name is not the same, expecting \" + expectedName + \", actual is \" + this.name + \".\");\n        }\n        if (!Object.is(this.count, expectedCount)) {\n            throw new Error(\"count is not the same, expecting \" + expectedCount + \", actual is \" + this.count + \".\");\n        }\n        expectedAllocation.forEach(function(expectedItem) {\n            var count = this.allocation.get(expectedItem.name);\n            if (!Object.is(count, expectedItem.count)) {\n                throw new Error(\"count of \" + expectedItem.name + \" is not the same, expecting \" + expectedItem.count + \", actual is \" + count + \".\");\n            }\n        }, this);\n    }\n};\nmodule.exports = SampleContract;";
source += x;

var dump_meta = function (meta) {
    for (var k in meta) {
        console.log('meta[' + k + ']=' + meta[k]);
        for (var kk in meta[k]) {
            console.log('meta[' + k + '][' + kk + ']=' + meta[k][kk]);
        }
    }
};

// var ast = esprima.parseScript(source, {
//     range: true,
//     loc: true
// }, function (node, meta) {
//     console.log('node.type=' + node.type);

//     var code = source.slice(meta.start.offset, meta.end.offset);
//     console.log('code is ' + code);

//     console.log('------------------');
// });
// console.log(ast);

var ast = esprima.parseScript(source, {
    range: true,
    loc: true
});

// console.log(ast);

function traverse(object, visitor, master, key) {
    var key, child, parent, path;

    parent = (typeof master === 'undefined') ? [] : master;

    if (visitor.call(null, object, parent, key) === false) {
        return;
    }
    for (key in object) {
        if (object.hasOwnProperty(key)) {
            child = object[key];
            path = [object];
            path.push.apply(path, parent);
            if (typeof child === 'object' && child !== null) {
                traverse(child, visitor, path, key);
            }
        }
    }
};

var flag = 1;

var exprStmts = new Map();

function InstructionInc_v01(val, path) {
    // console.log('path.len = ' + path.length);
    for (var i = 0; i < path.length; i++) {
        var item = path[i];
        if (!item.hasOwnProperty('type')) {
            console.log('typeof item is ' + typeof item);
            // console.log(item);
        } else {
            var node = item;
            console.log('node.type = ' + node.type);
            if (node.type == "ExpressionStatement") {
                // console.log('inst++');
                if (!node.hasOwnProperty('inst_flag')) {
                    node.inst_flag = "flag-" + flag;
                    flag++;
                    console.log('set flag ' + node.inst_flag);
                    exprStmts[node.inst_flag] = node;
                }
                if (!node.hasOwnProperty("inst_counter")) {
                    node.inst_counter = 0;
                }
                node.inst_counter += val;
                return;
            }
        }
    }
};

var ParentExpr = {
    ExpressionStatement: 1,
    BlockStatement: 1,
    IfStatement: 1,
    SwitchStatement: 1,
    DoWhileStatement: 1,
    ForStatement: 1,
    ForInStatement: 1,
    ForOfStatement: 1,
    WhileStatement: 1,
    WithStatement: 1,
    _XXX: 0
};

function InstructionInc(val, path) {
    for (var i = 0; i < path.length; i++) {
        var node = path[i];
        if (!node.hasOwnProperty('type')) {
            continue;
        }

        if (node.type in ParentExpr) {
            // console.log('inst++');
            if (!node.hasOwnProperty('inst_flag')) {
                node.inst_flag = "flag-" + flag;
                flag++;
                console.log('set flag ' + node.inst_flag);
                exprStmts[node.inst_flag] = node;
            }
            if (!node.hasOwnProperty("inst_counter")) {
                node.inst_counter = 0;
            }
            node.inst_counter += val;
            break;
        }
    }
};

var InstMap = {
    CallExpression: 1,
    AssignmentExpression: 1,
    BinaryExpression: 1,
    UpdateExpression: 1,
    UnaryExpression: 1,
    LogicalExpression: 1,
    _XXX: 0
};

traverse(ast, function (node, path, key) {
    if (node.hasOwnProperty('range')) {
        if (node.range[0] == 318 && node.range[1] == 334) {
            debugger;
            console.log('318-334: ' + path);
        }
    }
    if ((node.type in InstMap)) {
        // if (node.type == 'BinaryExpression') {
        //     for (var i = 0; i < path.length; i++) {
        //         console.log(i + ' : ' + typeof (path[i]));
        //     }
        // }
        console.log('--------------------');
        console.log('found expr ' + node.type + '; key is ' + key);
        InstructionInc(InstMap[node.type], path);
    } else if (node.type == 'FunctionExpression') {
        // console.log('FunctionExpression: node.params = ' + JSON.stringify(node.params));
    }
});

var new_source = "";
var start_offset = 0;
// traverse(ast, function (node, path) {
//     if (node.hasOwnProperty('range')) {
//         if (node.range[0] == 318 && node.range[1] == 335) {
//             debugger;
//         }
//     }
//     if (node.type == "ExpressionStatement") {
//         if (node.hasOwnProperty("inst_counter")) {
//             console.log('ExpressionStatement: {flag: ' + node.inst_flag + ', count: ' + node.inst_counter + '}.');
//             console.log('source scope: ' + node.range);
//             new_source += source.slice(start_offset, node.range[1]) + "\ninstruction_counter.incr('" + node.inst_flag + "', " + node.inst_counter + ");\n";
//             start_offset = node.range[1];
//         }
//     }
// });
// new_source += source.slice(start_offset);

var ranges = [];
for (var key in exprStmts) {
    ranges.push({
        offset: exprStmts[key].range[0],
        count: exprStmts[key].inst_counter,
        flag: exprStmts[key].inst_flag
    });
}
ranges.sort(function (a, b) {
    return a.offset - b.offset;
});
console.log(ranges);

ranges.forEach(function (item) {
    // console.log('!!!!!!!!!!!!!!!');
    // console.log('from {' + start_offset + ', ' + item.end_offset + '}');
    var c = source.slice(start_offset, item.offset);
    // console.log(c);
    new_source += source.slice(start_offset, item.offset);
    new_source += 'instruction_counter("' + item.flag + '",' + item.count + ');\n';
    start_offset = item.offset;
});
new_source += source.slice(start_offset);

console.log(new_source);

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

const module_path_prefix = (typeof process !== 'undefined') && (process.release.name === 'node') ? '../lib/' : '';

const instCounter = require(module_path_prefix + 'instruction_counter.js');

var source = "'use strict';\nvar doStatement = function() {\n    var i = 0;\n    do {\n        i++;\n    } while (i < 15);\n    do \n        i++;\n    while (i < 5);};\n\nvar SampleContract = function() {\n    LocalContractStorage.defineProperties(this, {\n        name: null,\n        count: null\n    });\n    LocalContractStorage.defineMapProperty(this, \"allocation\" + 123);\n    this.a = 0;\n    let foo = (s) => s+1;\n    new.target;\n    var elvisLives = Math.PI > 4 ? foo() : 'Nope';\n};\nSampleContract.prototype = {\n    init: function(name, count, allocation) {\n        this.name = name;\n        this.count = count;\n        this.zz[0] = 123;\n        allocation.forEach(function(item) {\n            this.allocation.put(item.name, item.count);\n        }, this);\n    },\n    dump: function() {\n        console.log('dump: this.name = ' + this.name);\n        console.log('dump: this.count = ' + this.count);\n        return this.a;\n    },\n    incr: function() {\n        this.a++;\n        var z = this.a;\n        console.log(this.dump());\n        return this.a;\n    },\n    verify: function(expectedName, expectedCount, expectedAllocation) {\n        if (!Object.is(this.name, expectedName))\n            throw new Error(\"name is not the same, expecting \" + expectedName + \", actual is \" + this.name + \".\");\n        else\n        var elvisLives = Math.PI > 4 ? foo() : 'Nope';\n\n        if (!Object.is(this.count, expectedCount)) {\n            throw new Error(\"count is not the same, expecting \" + expectedCount + \", actual is \" + this.count + \".\");\n        } else {\n            console.log('ok.');\n        }\n        expectedAllocation.forEach(function(expectedItem) {\n            var count = this.allocation.get(expectedItem.name);\n            if (!Object.is(count, expectedItem.count)) {\n                throw new Error(\"count of \" + expectedItem.name + \" is not the same, expecting \" + expectedItem.count + \", actual is \" + count + \".\");\n            }\n        }, this);\n    },\n    test_switch: function() {\n        var day = \"\";\n        switch (new Date().getDay()) {\n            case 0:\n                day = \"Sunday\";\n                break;\n            case 1:\n                day = \"Monday\";\n                break;\n            case 2:\n                day = \"Tuesday\";\n                break;\n            case 3:\n                day = \"Wednesday\";\n                break;\n            case 4:\n                day = \"Thursday\";\n                break;\n            case 5:\n                day = \"Friday\";\n                break;\n            case 6:\n                day = \"Saturday\";\n            default:\n                throw new Error(\"N/A\");\n        }\n        console.log('day is ' + day);\n    },\n    test_for: function() {\n        for (var i = 0; i < 123; i++) {\n            var z = i;\n            alert(z);\n        }\n        for (var i = 0; i < 123; i++)\n            alert(i);\n    }\n};\nmodule.exports = SampleContract;";

var new_source = instCounter.processScript(source);

console.log("\n" + new_source);

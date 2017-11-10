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

var esprima = require('esprima.js');

var source = "'use strict';var SampleContract = function () {LocalContractStorage.defineProperties(this, {name: null,count: null});LocalContractStorage.defineMapProperty(this, \"allocation\");};SampleContract.prototype = {init: function (name, count, allocation) {this.name = name;this.count = count;allocation.forEach(function (item) {this.allocation.put(item.name, item.count);}, this);},dump: function () {console.log('dump: this.name = ' + this.name);console.log('dump: this.count = ' + this.count);},verify: function (expectedName, expectedCount, expectedAllocation) {if (!Object.is(this.name, expectedName)) {throw new Error(\"name is not the same, expecting \" + expectedName + \", actual is \" + this.name + \".\");}if (!Object.is(this.count, expectedCount)) {throw new Error(\"count is not the same, expecting \" + expectedCount + \", actual is \" + this.count + \".\");}expectedAllocation.forEach(function (expectedItem) {var count = this.allocation.get(expectedItem.name);if (!Object.is(count, expectedItem.count)) {throw new Error(\"count of \" + expectedItem.name + \" is not the same, expecting \" + expectedItem.count + \", actual is \" + count + \".\");}}, this);}};module.exports = SampleContract;";

var ast = esprima.parseScript(source, {}, function (node, meta) {
    console.log('node.type=' + node.type);
    for (var k in meta) {
        console.log('meta[' + k + ']=' + meta[k]);
        for (var kk in meta[k]) {
            console.log('meta[' + k + '][' + kk + ']=' + meta[k][kk]);
        }
    }
    console.log('------------------');
});

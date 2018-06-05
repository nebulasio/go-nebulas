// Copyright (C) 2018 go-nebulas authors
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


var NebDate = (function(ProtoDate) {

    function NebDate() {
        if (!Blockchain) {
            throw new Error("'Blockchain' is not defined.");
        }
        if (!Blockchain.block) {
            throw new Error("'Blockchain.block' is not defined.");
        }
    
        var date = new(Function.prototype.bind.apply(ProtoDate, [ProtoDate].concat(Array.prototype.slice.call(arguments))))();
        if (arguments.length == 0) {
            // unit of timestamp is second
            date.setTime(Blockchain.block.timestamp * 1000);
        }
        Object.setPrototypeOf(date, NebDate.prototype);
        return date;
    }
    NebDate.now = function() {
        return new NebDate().getTime();
    }
    NebDate.UTC = function() {
        return ProtoDate.UTC.apply(null, Array.prototype.slice.call(arguments));
    }
    NebDate.parse = function(dateString) {
        return ProtoDate.parse(dateString);
    }

    NebDate.prototype.getYear = function() {
        throw new Error("Deprecated!");
    }
    NebDate.prototype.setYear = function() {
        throw new Error("Deprecated!");
    }

    NebDate.prototype.toLocaleDateString = function() {
        var tmp = new ProtoDate.prototype.constructor(this.getTime());
        return ProtoDate.prototype.toLocaleDateString.apply(tmp, Array.prototype.slice.call(arguments));
    }

    NebDate.prototype.toLocaleTimeString = function() {
        var tmp = new ProtoDate.prototype.constructor(this.getTime());
        return ProtoDate.prototype.toLocaleTimeString.apply(tmp, Array.prototype.slice.call(arguments));
    }

    NebDate.prototype.toLocaleString = function() {
        var tmp = new ProtoDate.prototype.constructor(this.getTime());
        return ProtoDate.prototype.toLocaleString.apply(tmp, Array.prototype.slice.call(arguments));
    }

    NebDate.prototype = new Proxy(NebDate.prototype, {
        getPrototypeOf: function(target) {
            throw new Error("Unsupported method!");
        },
    });

    Object.setPrototypeOf(NebDate.prototype, ProtoDate.prototype);
    return NebDate;
})(Date);

module.exports = NebDate;
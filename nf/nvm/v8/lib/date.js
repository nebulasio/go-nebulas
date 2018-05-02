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


var NebDate = (function(Date) {
    function NebDate() {
        if (!Blockchain) {
            throw new Error("'Blockchain' is not defined.");
        }
        if (!Blockchain.block) {
            throw new Error("'Blockchain.block' is not defined.");
        }
    
        var date = new(Function.prototype.bind.apply(Date, [Date].concat(Array.prototype.slice.call(arguments))))();
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
        return Date.UTC.apply(null, arguments);
    }
    NebDate.parse = function(dateString) {
        return Date.parse(dateString);
    }
    NebDate.prototype = new Proxy(NebDate.prototype, {
        getPrototypeOf: function(target) {
            throw new Error("Unsupported method!");
        },
    });
    Object.setPrototypeOf(NebDate.prototype, Date.prototype);
    return NebDate;
})(Date);

module.exports = NebDate;
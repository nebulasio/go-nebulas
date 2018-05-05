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
    // compatibility
    var Date = function() {
        throw new Error("Date is not allowed in nvm.");
    }

    function allow() {
        return Blockchain.block.seed != null && typeof(Blockchain.block.seed) !== 'undefined';
    }

    function NebDate() {
        if (!Blockchain) {
            throw new Error("'Blockchain' is not defined.");
        }
        if (!Blockchain.block) {
            throw new Error("'Blockchain.block' is not defined.");
        }
        if (!allow()) {
            throw new Error("Date is not allowed in nvm.");
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
        if (!allow()) {
            Date.now();
        }
        return new NebDate().getTime();
    }
    NebDate.UTC = function() {
        if (!allow()) {
            Date.UTC();
        }
        return ProtoDate.UTC.apply(null, arguments);
    }
    NebDate.parse = function(dateString) {
        if (!allow()) {
            Date.parse(dateString);
        }
        return ProtoDate.parse(dateString);
    }

    NebDate.prototype.getTimezoneOffset = function() {
        throw new Error("Unsupported method!");
    }
    NebDate.prototype.getDate = function() {
        return this.getUTCDate();
    }
    NebDate.prototype.getDay = function() {
        return this.getUTCDay();
    }
    NebDate.prototype.getFullYear = function() {
        return this.getUTCFullYear();
    }
    NebDate.prototype.getHours = function() {
        return this.getUTCHours();
    }
    NebDate.prototype.getMilliseconds = function() {
        return this.getUTCMilliseconds();
    }
    NebDate.prototype.getMinutes = function() {
        return this.getUTCMinutes();
    }
    NebDate.prototype.getMonth = function() {
        return this.getUTCMonth();
    }
    NebDate.prototype.getSeconds = function() {
        return this.getUTCSeconds();
    },
    NebDate.prototype.getYear = function() {
        throw new Error("Deprecated!");
    }
    NebDate.prototype.setYear = function() {
        throw new Error("Deprecated!");
    }
    NebDate.prototype.setDate = function() {
        return this.setUTCDate.apply(this, arguments);
    }
    NebDate.prototype.setFullYear = function() {
        return this.setUTCFullYear.apply(this, arguments);
    }
    NebDate.prototype.setHours = function() {
        return this.setUTCHours.apply(this, arguments);
    }
    NebDate.prototype.setMilliseconds = function() {
        return this.setUTCMilliseconds.apply(this, arguments);
    }
    NebDate.prototype.setMinutes = function() {
        return this.setUTCMinutes.apply(this, arguments);
    }
    NebDate.prototype.setMonth = function() {
        return this.setUTCMonth.apply(this, arguments);
    }
    NebDate.prototype.setSeconds = function() {
        return this.setUTCSeconds.apply(this, arguments);
    }
    NebDate.prototype.toString = function() {
        // return UTC string
        return this.toUTCString.apply(this, arguments);
    }
    NebDate.prototype.toDateString = function() {
        throw new Error("Unsupported method!");
    }
    NebDate.prototype.toTimeString = function() {
        throw new Error("Unsupported method!");
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
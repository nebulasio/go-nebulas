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

var StandardToken = function () {
    LocalContractStorage.defineProperties(this, {
        name: null,
        symbol: null,
        _totalSupply: null,
        totalIssued: null
    });
    LocalContractStorage.defineMapProperty(this, "balances");
};

StandardToken.prototype = {
    init: function (name, symbol, totalSupply) {
        this.name = name;
        this.symbol = symbol;
        this._totalSupply = totalSupply;
        this.totalIssued = 0;
        throw 'fail to init';
    },
    totalSupply: function () {
        return this._totalSupply;
    },
    balanceOf: function (owner) {
        return this.balances.get(owner) || 0;
    },
    transfer: function (to, value) {
        var balance = this.balanceOf(msg.sender);
        if (balance < value) {
            return false;
        }

        var finalBalance = balance - value;
        this.balances.set(msg.sender, finalBalance);
        this.balances.set(to, this.balanceOf(to) + value);
        return true;
    },
    pay: function (msg, amount) {
        if (this.totalIssued + amount > this._totalSupply) {
            throw new Error("too much amount, exceed totalSupply");
        }
        this.balances.set(msg.sender, this.balanceOf(msg.sender) + amount);
        this.totalIssued += amount;
    }
};

module.exports = StandardToken;

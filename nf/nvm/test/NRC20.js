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

let Allowed = function (obj) {
    this._allowed = {};
    this.parse(obj);
};

Allowed.prototype = {
    toString: function () {
        return JSON.stringify(this._allowed);
    },

    parse: function (obj) {
        if (typeof obj != "undefined") {
            let data = JSON.parse(obj);
            for (let key in data) {
                this._allowed[key] = new BigNumber(data[key]);
            }
        }
    },

    get: function (key) {
        return this._allowed[key];
    },

    set: function (key, value) {
        this._allowed[key] = new BigNumber(value);
    }
};

let StandardToken = function () {
    LocalContractStorage.defineProperties(this, {
        _name: null,
        _symbol: null,
        _decimals: null,
        _totalSupply: {
            parse: function (value) {
                return new BigNumber(value);
            },
            stringify: function (o) {
                return o.toString(10);
            }
        }
    });

    LocalContractStorage.defineMapProperties(this, {
        "_balances": {
            parse: function (value) {
                return new BigNumber(value);
            },
            stringify: function (o) {
                return o.toString(10);
            }
        },
        "_allowed": {
            parse: function (value) {
                return new Allowed(value);
            },
            stringify: function (o) {
                return o.toString();
            }
        }
    });
};

StandardToken.prototype = {
    init: function (name, symbol, decimals, totalSupply) {
        this._name = name;
        this._symbol = symbol;
        this._decimals = decimals || 0;
        this._totalSupply = new BigNumber(totalSupply).mul(new BigNumber(10).pow(decimals));

        let from = Blockchain.transaction.from;
        this._balances.set(from, this._totalSupply);
        this._transferEvent(true, from, from, this._totalSupply);
    },

    // Returns the name of the token
    name: function () {
        return this._name;
    },

    // Returns the symbol of the token
    symbol: function () {
        return this._symbol;
    },

    // Returns the number of decimals the token uses
    decimals: function () {
        return this._decimals;
    },

    totalSupply: function () {
        return this._totalSupply.toString(10);
    },

    balanceOf: function (owner) {
        this._verifyAddress(owner);

        let balance = this._balances.get(owner);
        if (balance instanceof BigNumber) {
            return balance.toString(10);
        } else {
            return "0";
        }
    },
    _verifyAddress: function (address) {
        if (Blockchain.verifyAddress(address) === 0) {
            throw new Error("Address format error, address=" + address);
        }
    },

    _verifyValue: function(value) {
        let bigVal = new BigNumber(value);
        if (bigVal.isNaN() || !bigVal.isFinite()) {
            throw new Error("Invalid value, value=" + value);
        }
        if (bigVal.isNegative()) {
            throw new Error("Value is negative, value=" + value);
        }
        if (!bigVal.isInteger()) {
            throw new Error("Value is not integer, value=" + value);
        }
        if (value !== bigVal.toString(10)) {
            throw new Error("Invalid value format.");
        }
    },

    transfer: function (to, value) {
        this._verifyAddress(to);
        this._verifyValue(value);

        value = new BigNumber(value);
        let from = Blockchain.transaction.from;
        let balance = this._balances.get(from) || new BigNumber(0);

        if (balance.lt(value)) {
            throw new Error("transfer failed.");
        }

        this._balances.set(from, balance.sub(value));
        let toBalance = this._balances.get(to) || new BigNumber(0);
        this._balances.set(to, toBalance.add(value));

        this._transferEvent(true, from, to, value.toString(10));
    },

    transferFrom: function (from, to, value) {
        this._verifyAddress(from);
        this._verifyAddress(to);
        this._verifyValue(value);

        let spender = Blockchain.transaction.from;
        let balance = this._balances.get(from) || new BigNumber(0);

        let allowed = this._allowed.get(from) || new Allowed();
        let allowedValue = allowed.get(spender) || new BigNumber(0);
        value = new BigNumber(value);

        if (balance.gte(value) && allowedValue.gte(value)) {

            this._balances.set(from, balance.sub(value));

            // update allowed value
            allowed.set(spender, allowedValue.sub(value));
            this._allowed.set(from, allowed);

            let toBalance = this._balances.get(to) || new BigNumber(0);
            this._balances.set(to, toBalance.add(value));

            this._transferEvent(true, from, to, value.toString(10));
        } else {
            throw new Error("transfer failed.");
        }
    },

    _transferEvent: function (status, from, to, value) {
        Event.Trigger(this.name(), {
            Status: status,
            Transfer: {
                from: from,
                to: to,
                value: value
            }
        });
    },

    approve: function (spender, currentValue, value) {
        this._verifyAddress(spender);
        this._verifyValue(currentValue);
        this._verifyValue(value);

        let from = Blockchain.transaction.from;

        let oldValue = this.allowance(from, spender);
        if (oldValue != currentValue) {
            throw new Error("current approve value mistake.");
        }

        let balance = new BigNumber(this.balanceOf(from));
        value = new BigNumber(value);

        if (balance.lt(value)) {
            throw new Error("invalid value.");
        }

        let owned = this._allowed.get(from) || new Allowed();
        owned.set(spender, value);

        this._allowed.set(from, owned);

        this._approveEvent(true, from, spender, value.toString(10));
    },

    _approveEvent: function (status, from, spender, value) {
        Event.Trigger(this.name(), {
            Status: status,
            Approve: {
                owner: from,
                spender: spender,
                value: value
            }
        });
    },

    allowance: function (owner, spender) {
        this._verifyAddress(owner);
        this._verifyAddress(spender);

        let owned = this._allowed.get(owner);
        if (owned instanceof Allowed) {
            let spenderObj = owned.get(spender);
            if (typeof spenderObj != "undefined") {
                return spenderObj.toString(10);
            }
        }
        return "0";
    }
};

module.exports = StandardToken;
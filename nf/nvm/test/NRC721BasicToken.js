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

var Operator = function (obj) {
    this.operator = {};
    this.parse(obj);
};

Operator.prototype = {
    toString: function () {
        return JSON.stringify(this.operator);
    },

    parse: function (obj) {
        if (typeof obj != "undefined") {
            var data = JSON.parse(obj);
            for (var key in data) {
                this.operator[key] = data[key];
            }
        }
    },

    get: function (key) {
        return this.operator[key];
    },

    set: function (key, value) {
        this.operator[key] = value;
    }
};

var StandardToken = function () {
    LocalContractStorage.defineProperties(this, {
        _name: null,
        _symbol: null,
    });

    LocalContractStorage.defineMapProperties(this, {
        "tokenOwner": null,
        "ownedTokensCount": {
            parse: function (value) {
                return new BigNumber(value);
            },
            stringify: function (o) {
                return o.toString(10);
            }
        },
        "tokenApprovals": null,
        "operatorApprovals": {
            parse: function (value) {
                return new Operator(value);
            },
            stringify: function (o) {
                return o.toString();
            }
        },
        
    });
};

StandardToken.prototype = {
    init: function (name, symbol) {
        this._name = name;
        this._symbol = symbol;
    },

    name: function () {
        return this._name;
    },

    symbol: function () {
        return this._symbol;
    },

    balanceOf: function (_owner) {
        var balance = this.ownedTokensCount.get(_owner);
        if (balance instanceof BigNumber) {
            return balance.toString(10);
        } else {
            return "0";
        }
    },

    ownerOf: function (_tokenId) {
        return this.tokenOwner.get(_tokenId);
    },

    approve: function (_to, _tokenId) {
        var from = Blockchain.transaction.from;

        var owner = this.ownerOf(_tokenId);
        if (_to == owner) {
            throw new Error("invalid address in approve.");
        }
        // msg.sender == owner || isApprovedForAll(owner, msg.sender)
        if (owner == from || this.isApprovedForAll(owner, from)) {
            this.tokenApprovals.set(_tokenId, _to);
            this.approveEvent(true, owner, _to, _tokenId);
        } else {
            throw new Error("permission denied in approve.");
        }
    },

    getApproved: function (_tokenId) {
        return this.tokenApprovals.get(_tokenId);
    },

    setApprovalForAll: function(_to, _approved) {
        var from = Blockchain.transaction.from;
        if (from == _to) {
            throw new Error("invalid address in setApprovalForAll.");
        }
        var operator = this.operatorApprovals.get(from) || new Operator();
        operator.set(_to, _approved);
        this.operatorApprovals.set(from, operator);
    },

    isApprovedForAll: function(_owner, _operator) {
        var operator = this.operatorApprovals.get(_owner);
        if (operator != null) {
            if (operator.get(_operator) === "true") {
                return true;
            } else {
                return false;
            }
        }
    },

    isApprovedOrOwner: function(_spender, _tokenId) {
        var owner = this.ownerOf(_tokenId);
        return _spender == owner || this.getApproved(_tokenId) == _spender || this.isApprovedForAll(owner, _spender);
    },

    transferFrom: function (_from, _to, _tokenId) {
        var from = Blockchain.transaction.from;
        if (this.isApprovedOrOwner(from, _tokenId)) {
            this.clearApproval(_from, _tokenId);
            this.removeTokenFrom(_from, _tokenId);
            this.addTokenTo(_to, _tokenId);
            this.transferEvent(true, _from, _to, _tokenId);
        } else {
            throw new Error("permission denied in transferFrom.");
        }
        
    },

    clearApproval: function (_owner, _tokenId) {
        var owner = this.ownerOf(_tokenId);
        if (_owner != owner) {
            throw new Error("permission denied in clearApproval.");
        }
        this.tokenApprovals.del(_tokenId);
    },

    removeTokenFrom: function(_from, _tokenId) {
        if (_from != this.ownerOf(_tokenId)) {
            throw new Error("permission denied in removeTokenFrom.");
        }
        var tokenCount = this.ownedTokensCount.get(_from);
        if (tokenCount.lt(1)) {
            throw new Error("Insufficient account balance in removeTokenFrom.");
        }
        this.ownedTokensCount.set(_from, tokenCount-1);
    },

    addTokenTo: function(_to, _tokenId) {
        this.tokenOwner.set(_tokenId, _to);
        var tokenCount = this.ownedTokensCount.get(_to) || new BigNumber(0);
        this.ownedTokensCount.set(_to, tokenCount+1);
    },

    mint: function(_to, _tokenId) {
        this.addTokenTo(_to, _tokenId);
        this.transferEvent(true, "", _to, _tokenId);
    },

    burn: function(_owner, _tokenId) {
        this.clearApproval(_owner, _tokenId);
        this.removeTokenFrom(_owner, _tokenId);
        this.transferEvent(true, _owner, "", _tokenId);
    },

    transferEvent: function (status, _from, _to, _tokenId) {
        Event.Trigger(this.name(), {
            Status: status,
            Transfer: {
                from: _from,
                to: _to,
                tokenId: _tokenId
            }
        });
    },

    approveEvent: function (status, _owner, _spender, _tokenId) {
        Event.Trigger(this.name(), {
            Status: status,
            Approve: {
                owner: _owner,
                spender: _spender,
                tokenId: _tokenId
            }
        });
    }

};

module.exports = StandardToken;

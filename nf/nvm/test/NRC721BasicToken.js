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
    init: function (name) {
        this._name = name;
    },

    name: function () {
        return this._name;
    },

    // Returns the number of tokens owned by owner.
    balanceOf: function (owner) {
        var balance = this.ownedTokensCount.get(owner);
        if (balance instanceof BigNumber) {
            return balance.toString(10);
        } else {
            return "0";
        }
    },

    //Returns the address of the owner of the tokenID.
    ownerOf: function (tokenID) {
        return this.tokenOwner.get(tokenID);
    },

    /**
     * Set or reaffirm the approved address for an token.
     * The function SHOULD throws unless transcation from is the current token owner, or an authorized operator of the current owner.
     */
    approve: function (to, tokenId) {
        var from = Blockchain.transaction.from;

        var owner = this.ownerOf(tokenId);
        if (to == owner) {
            throw new Error("invalid address in approve.");
        }
        if (owner == from || this.isApprovedForAll(owner, from)) {
            this.tokenApprovals.set(tokenId, to);
            this._approveEvent(true, owner, to, tokenId);
        } else {
            throw new Error("permission denied in approve.");
        }
    },

    // Returns the approved address for a single token.
    getApproved: function (tokenId) {
        return this.tokenApprovals.get(tokenId);
    },

    /**
     * Enable or disable approval for a third party (operator) to manage all of transaction from's assets.
     * operator Address to add to the set of authorized operators. 
     * @param approved True if the operators is approved, false to revoke approval
     */
    setApprovalForAll: function(to, approved) {
        var from = Blockchain.transaction.from;
        if (from == to) {
            throw new Error("invalid address in setApprovalForAll.");
        }
        var operator = this.operatorApprovals.get(from) || new Operator();
        operator.set(to, approved);
        this.operatorApprovals.set(from, operator);
    },

    /**
     * @dev Tells whether an operator is approved by a given owner
     * @param owner owner address which you want to query the approval of
     * @param operator operator address which you want to query the approval of
     * @return bool whether the given operator is approved by the given owner
     */
    isApprovedForAll: function(owner, operator) {
        var operator = this.operatorApprovals.get(owner);
        if (operator != null) {
            if (operator.get(operator) === "true") {
                return true;
            } else {
                return false;
            }
        }
    },


    /**
     * @dev Returns whether the given spender can transfer a given token ID
     * @param spender address of the spender to query
     * @param tokenId uint256 ID of the token to be transferred
     * @return bool whether the msg.sender is approved for the given token ID,
     *  is an operator of the owner, or is the owner of the token
     */
    _isApprovedOrOwner: function(spender, tokenId) {
        var owner = this.ownerOf(tokenId);
        return spender == owner || this.getApproved(tokenId) == spender || this.isApprovedForAll(owner, spender);
    },

    /**
     * Transfers the ownership of an token from one address to another address. 
     * The caller is responsible to confirm that to is capable of receiving token or else they may be permanently lost.
     * Transfers tokenId from address from to address to, and MUST fire the Transfer event.
     * The function SHOULD throws unless the transaction from is the current owner, an authorized operator, or the approved address for this token. 
     * Throws if from is not the current owner. 
     * Throws if to is the contract address. 
     * Throws if tokenId is not a valid token.
     */
    transferFrom: function (from, to, tokenId) {
        var sender = Blockchain.transaction.from;
        var contractAddress = Blockchain.transaction.to;
        if (contractAddress == to) {
            throw new Error("Forbidden to transfer money to a smart contract address");
        }
        if (this._isApprovedOrOwner(sender, tokenId)) {
            this._clearApproval(from, tokenId);
            this._removeTokenFrom(from, tokenId);
            this._addTokenTo(to, tokenId);
            this._transferEvent(true, from, to, tokenId);
        } else {
            throw new Error("permission denied in transferFrom.");
        }
        
    },


     /**
     * Internal function to clear current approval of a given token ID
     * Throws if the given address is not indeed the owner of the token
     * @param sender owner of the token
     * @param tokenId uint256 ID of the token to be transferred
     */
    _clearApproval: function (sender, tokenId) {
        var owner = this.ownerOf(tokenId);
        if (sender != owner) {
            throw new Error("permission denied in clearApproval.");
        }
        this.tokenApprovals.del(tokenId);
    },

    /**
     * Internal function to remove a token ID from the list of a given address
     * @param from address representing the previous owner of the given token ID
     * @param tokenId uint256 ID of the token to be removed from the tokens list of the given address
     */
    _removeTokenFrom: function(from, tokenId) {
        if (from != this.ownerOf(tokenId)) {
            throw new Error("permission denied in removeTokenFrom.");
        }
        var tokenCount = this.ownedTokensCount.get(from);
        if (tokenCount.lt(1)) {
            throw new Error("Insufficient account balance in removeTokenFrom.");
        }
        this.ownedTokensCount.set(from, tokenCount.sub(1));
    },

    /**
     * Internal function to add a token ID to the list of a given address
     * @param to address representing the new owner of the given token ID
     * @param tokenId uint256 ID of the token to be added to the tokens list of the given address
     */
    _addTokenTo: function(to, tokenId) {
        this.tokenOwner.set(tokenId, to);
        var tokenCount = this.ownedTokensCount.get(to) || new BigNumber(0);
        this.ownedTokensCount.set(to, tokenCount.add(1));
    },

    /**
     * Internal function to mint a new token
     * @param to The address that will own the minted token
     * @param tokenId uint256 ID of the token to be minted by the msg.sender
     */
    _mint: function(to, tokenId) {
        this._addTokenTo(to, tokenId);
        this._transferEvent(true, "", to, tokenId);
    },

    /**
     * Internal function to burn a specific token
     * @param tokenId uint256 ID of the token being burned by the msg.sender
     */
    _burn: function(owner, tokenId) {
        this._clearApproval(owner, tokenId);
        this._removeTokenFrom(owner, tokenId);
        this._transferEvent(true, owner, "", tokenId);
    },

    _transferEvent: function (status, from, to, tokenId) {
        Event.Trigger(this.name(), {
            Status: status,
            Transfer: {
                from: from,
                to: to,
                tokenId: tokenId
            }
        });
    },

    _approveEvent: function (status, owner, spender, tokenId) {
        Event.Trigger(this.name(), {
            Status: status,
            Approve: {
                owner: owner,
                spender: spender,
                tokenId: tokenId
            }
        });
    }

};

module.exports = StandardToken;

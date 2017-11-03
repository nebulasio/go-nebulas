'use strict';

var StandardToken = function () {
    LocalContractStorage.defineProperties(this, "name", "symbol", "totalSupply", "totalIssued");
    LocalContractStorage.defineMapProperties(this, "balances");
};

StandardToken.prototype = {
    init: function (name, symbol, totalSupply) {
        this.name = name;
        this.symbol = symbol;
        this.totalSupply = totalSupply;
        this.totalIssued = 0;
    },
    totalSupply: function () {
        return this.totalSupply;
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
    pay: function (amount) {
        if (this.totalIssued + amount > this.totalSupply) {
            throw new Error("too much amount, exceed totalSupply");
        }

        this.balances.set(msg.sender, this.balanceOf(msg.sender) + amount);
        this.totalIssued += amount;
    }
};

module.exports = StandardToken;

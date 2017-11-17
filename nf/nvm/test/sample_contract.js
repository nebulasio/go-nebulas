'use strict';

var SampleContract = function () {
    LocalContractStorage.defineProperties(this, {
        name: null,
        count: null
    });
    LocalContractStorage.defineMapProperty(this, "allocation");
};

SampleContract.prototype = {
    init: function (name, count, allocation) {
        this.name = name;
        this.count = count;
        allocation.forEach(function (item) {
            this.allocation.put(item.name, item.count);
        }, this);
        // console.log('init: this.name = ' + this.name);
        // console.log('init: this.count = ' + this.count);
        console.log('init: Blockchain.current.coinbase = ' + Blockchain.current.coinbase);
        console.log('init: Blockchain.current.blockNonce = ' + Blockchain.current.blockNonce);
        console.log('init: Blockchain.current.blockHash = ' + Blockchain.current.blockHash);
        console.log('init: Blockchain.current.blockHeight = ' + Blockchain.current.blockHeight);
        console.log('init: Blockchain.current.txNonce = ' + Blockchain.current.txNonce);
        console.log('init: Blockchain.current.txHash = ' + Blockchain.current.txHash);
    },
    dump: function () {
        console.log('dump: this.name = ' + this.name);
        console.log('dump: this.count = ' + this.count);
    },
    verify: function (expectedName, expectedCount, expectedAllocation) {
        if (!Object.is(this.name, expectedName)) {
            throw new Error("name is not the same, expecting " + expectedName + ", actual is " + this.name + ".");
        }
        if (!Object.is(this.count, expectedCount)) {
            throw new Error("count is not the same, expecting " + expectedCount + ", actual is " + this.count + ".");
        }

        expectedAllocation.forEach(function (expectedItem) {
            var count = this.allocation.get(expectedItem.name);
            if (!Object.is(count, expectedItem.count)) {
                throw new Error("count of " + expectedItem.name + " is not the same, expecting " + expectedItem.count + ", actual is " + count + ".");
            }
        }, this);
    }
};

module.exports = SampleContract;

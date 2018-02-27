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
        console.log('init: Blockchain.block.coinbase = ' + Blockchain.block.coinbase);
        console.log('init: Blockchain.block.hash = ' + Blockchain.block.hash);
        console.log('init: Blockchain.block.height = ' + Blockchain.block.height);
        console.log('init: Blockchain.transaction.from = ' + Blockchain.transaction.from);
        console.log('init: Blockchain.transaction.to = ' + Blockchain.transaction.to);
        console.log('init: Blockchain.transaction.value = ' + Blockchain.transaction.value);
        console.log('init: Blockchain.transaction.nonce = ' + Blockchain.transaction.nonce);
        console.log('init: Blockchain.transaction.hash = ' + Blockchain.transaction.hash);
    },
    dump: function () {
        console.log('dump: this.name = ' + this.name);
        console.log('dump: this.count = ' + this.count);
    },
    $dump: function () {
        return this.dump();
    },
    dump_1: function () {
        return this.dump();
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

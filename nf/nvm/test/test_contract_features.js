'use strict';

var testContract = function() {

}

var assert = function(expression, info) {
    if (!expression) {
        throw info;
    }
};

testContract.prototype = {
    init: function() {
        
    },

    testGetAccountState: function() {
        return Blockchain.getAccountState(Blockchain.transaction.from).balance;
    },

    testGetAccountStateWrongAddr: function() {
        return Blockchain.getAccountState("n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR").balance;
    },

    testGetAccountState1: function() {
        var from = Blockchain.transaction.from;
        var to = Blockchain.transaction.to;
        var value = new BigNumber(Blockchain.transaction.value);

        var fromState1 = Blockchain.getAccountState(from);
        // console.log("============", JSON.stringify(fromState1));
        var toState1 = Blockchain.getAccountState(to);
        // console.log(JSON.stringify(toState1));

        Blockchain.transfer(from, value);
        var fromState2 = Blockchain.getAccountState(from);
        // console.log(JSON.stringify(fromState2));

        var toState2 = Blockchain.getAccountState(to);
        // console.log(JSON.stringify(toState2));
        var fromBalance1 = new BigNumber(fromState1.balance);
        var fromBalance2 = new BigNumber(fromState2.balance);
        console.log(fromBalance1);
        console.log(fromBalance2);
        assert(fromBalance1.add(value).equals(fromBalance2), "err 1");

        var toBalance1 = new BigNumber(toState1.balance);
        var toBalance2 = new BigNumber(toState2.balance);
        assert(toBalance1.sub(value).equals(toBalance2), "err 2");
        assert(parseInt(toState1.nonce) == parseInt(toState2.nonce), "err3");
    },

    testGetAccountState2: function() {
        return Blockchain.getAccountState("0x1233455");
    },

    testGetPreBlockHash: function(offset) {
        var hash = Blockchain.getPreBlockHash(offset);
        var height = Blockchain.block.height;
        return {hash: hash, height: height};
    },
    testGetPreBlockHash1: function(offset) {
        return  Blockchain.getPreBlockHash(offset);
    },

    testGetPreBlockSeed: function(offset) {
        var seed = Blockchain.getPreBlockSeed(offset);
        var height = Blockchain.block.height;
        return {seed: seed, height: height};
    },

    testGetPreBlockSeed1: function(offset) {
        return Blockchain.getPreBlockSeed(offset);
    },

    testGetPreBlockHashByNativeBlock: function(offset) {
        return Blockchain.nativeBlockchain.getPreBlockHash(offset);
    },

    testGetPreBlockSeedByNativeBlock: function(offset) {
        return Blockchain.nativeBlockchain.getPreBlockSeed(offset);
    }

}

module.exports = testContract;
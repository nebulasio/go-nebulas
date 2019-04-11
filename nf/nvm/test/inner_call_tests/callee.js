"use strict";

function Contract() {}

Contract.prototype = {
	init: function () {
        //
    },

    // this function returns true if deployed in or after version 1.1.0, otherwise false
    supportInnerCall: function() {
        return typeof(Blockchain.Contract);
    },

    supportRandom: function() {
        return {
            randseed: typeof(Math.random.seed) === 'function',
            blockseed: Blockchain.block.seed != ''
        };
    }
};

module.exports = Contract;
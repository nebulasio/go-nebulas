"use strict";

function Contract() {}
// deploy this contract in or after 1.1.0
Contract.prototype = {
	init: function () {
        //
    },

    testInnerCall: function(address, expectResult) {
        var funcs =  {
            supportInnerCall: function() { 
                
            }
        }
        
        var c = new Blockchain.Contract(address, funcs);

        var actual = c.value(0).call("supportInnerCall");
        if (expectResult != actual) {
            throw new Error("expect: " + expectResult + ", actual: " + actual);
        }
    },

    testRandom: function(address, expRandseed, expBlockseed, expRandseedInCaller, expBlockseedInCaller) {
        var funcs =  {
            supportRandom: function() { 
                
            }
        }

        var c = new Blockchain.Contract(address, funcs);

        var actual = c.value(0).call("supportRandom");
        console.error("actual:", actual);
        if (expRandseed != actual.randseed) {
            throw  new Error("expRandseed: " + expRandseed + ", actual: " + actual.randseed);
        }
        if (expBlockseed != actual.blockseed) {
            throw new Error("expBlockseed: " + expBlockseed + ", actual: " + actual.blockseed);
        }
        if (expRandseedInCaller != (typeof(Math.random.seed) === 'function')) {
            throw new Error("expect: Math.random.seed exists in caller ? " + expRandseedInCaller);
        }
        if (expBlockseedInCaller != Blockchain.block.seed != '') {
            throw  new Error("expect: Blockchain.block.seed=" + expBlockseedInCaller);
        }
    }
};

module.exports = Contract;

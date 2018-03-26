'use strict'

var TransferValueContract = function () {
    // LocalContractStorge.defineProperties(this, {
    //     totalBalance: null
    // })
}


TransferValueContract.prototype = {
     init: function() {
    //     this.totalBalance = 0;
     },

    transfer: function(to) {
        var result = Blockchain.transfer(to, Blockchain.transaction.value);
        // var result = Blockchain.transfer(to, 0);
        if (!result) {
	    	throw new Error("transfer failed.");
        }
        return Blockchain.transaction.value;
    },
    
}
module.exports = TransferValueContract;
"use strict";


var ProxyBankContract = function () {

};

// save value to contract, only after height of block, users can takeout
ProxyBankContract.prototype = {
	init: function () {
        //
    },

    callWhile: function() {
        console.error("callee ....... ");
        var i = 0;
        while (i < 10) {
            i++;
        }
        return 0;
    },

};

module.exports = ProxyBankContract;
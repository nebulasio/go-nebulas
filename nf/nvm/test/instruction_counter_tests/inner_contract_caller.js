"use strict";


var ProxyBankContract = function () {

};

// save value to contract, only after height of block, users can takeout
ProxyBankContract.prototype = {
	init: function () {
        //
    },

    callWhile: function(address, to) {
        var funcs =  {
            callWhile: function() { 
                
            }
        }
        
        var c = new Blockchain.Contract(address, funcs);

        // var args = 
        // var args = "[\"" + to + "\"]";
        console.error("caller ....... ", address, to);
        c.value(0).call("callWhile", to); 
        return "";
    },

};

module.exports = ProxyBankContract;

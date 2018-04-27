"use strict";


var ProxyBankContract = function () {

};

// save value to contract, only after height of block, users can takeout
ProxyBankContract.prototype = {
	init: function () {
        //
    },
    
    save: function (address, height) {

        var funcs =  {
            save: function() { 
            }
        }

        var c = new Blockchain.Contract("n21223", funcs);


        c.value(5).save(height); 
    }
};

module.exports = ProxyBankContract;

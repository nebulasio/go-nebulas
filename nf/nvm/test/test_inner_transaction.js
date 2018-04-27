"use strict";


var ProxyBankContract = function () {

};

// save value to contract, only after height of block, users can takeout
ProxyBankContract.prototype = {
	init: function () {
        //
    },
    
    save: function (address, to, height) {

        var funcs =  {
            save: function() { 
            
            }
        }

        var c = new Blockchain.Contract(address, funcs);

        var args = "[\"" + to + "\", \""+ height +"\"]";
        c.value(5).call("save", args); 
        this.transferEvent(true, address, height);
    },
    transferEvent: function (status, address, height) {
        Event.Trigger("test_inner_transaction", {
            Status: status,
            Transfer: {
                address: address,
                height: height,
                magic: "main"
            }
        });
    },
};

module.exports = ProxyBankContract;

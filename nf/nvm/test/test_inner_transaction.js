"use strict";


var ProxyBankContract = function () {

};

// save value to contract, only after height of block, users can takeout
ProxyBankContract.prototype = {
	init: function () {
        //
    },
    getRandom: function(address, to) {
        var funcs =  {
            getRandom: function() { 
                
            }
        }
        
        var rand = _native_math.random();
        console.log("rand:", rand);

        var c = new Blockchain.Contract(address, funcs);

        // var args = 
        // var args = "[\"" + to + "\"]";
        c.value(0).call("getRandom", to); 
        return rand;
    },
    getRandomSingle: function(address, to) {
        var funcs =  {
            getRandom: function() { 
                
            }
        }
        
        var rand = _native_math.random();
        console.log("rand:", rand);
        return rand;
    },
    getSource: function(address) {
        var funcs =  {
            save: function() { 
                
            }
        }
        var c = new Blockchain.Contract(address, funcs);
        console.log("=============================address:", address);
        // return c;
        // c.value(5).call("save", args); 
    },
    save: function (address, to, height) {

        console.log("enter inner transaction begin create inner client");
        var c = new Blockchain.Contract(address);
        console.log("exit inner transaction");
        // var args = "[\"" + to + "\", \""+ height +"\"]";
        c.value(5).call("save", to, height); 
        this.transferEvent(true, address, height);
    },
    saveMem: function (address, to, mem) {
        var funcs =  {
            saveMem: function() { 
            
            }
        }
        console.log("saveMem:", mem);
        var m = new ArrayBuffer(mem);
        var c = new Blockchain.Contract(address, funcs);
        
        // var args = "[\"" + to + "\", \""+ mem +"\"]";
        c.value(0).call("saveMem", to, mem); 
        this.transferEvent(true, address, 0, mem);
    },
    saveErr: function(address, to, flag) {
        if (flag == 0) {
            throw("saveErr in test_inner_transaction");
            return;
        }
        var funcs =  {
            saveErr: function() { 
            
            }
        }
        var c = new Blockchain.Contract(address, funcs);
        // var args = "[\"" + to + "\", \""+ flag +"\"]";
        c.value(0).call("saveErr", to, flag); 
        // this.transferEvent(true, address, 0, mem);
    },
    saveToLoop: function (address, to, height) {
        var c = new Blockchain.Contract(address);

        // var args = "[\"" + address + "\", \"" + to + "\", \""+ height +"\"]";
        c.value(5).call("saveToLoop", address, to, height); 
        this.transferEvent(true, address, height);
    },
    transferEvent: function (status, address, height, mem) {
        Event.Trigger("test_inner_transaction", {
            Status: status,
            Transfer: {
                address: address,
                height: height,
                mem: mem,
                magic: "main"
            }
        });
    },

};

module.exports = ProxyBankContract;

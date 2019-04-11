"use strict"

var kvStore = function() {
}

kvStore.prototype = {
    init: function() {
        //
    },

    testTimeOut: function(address) {
        var real_kv = new Blockchain.Contract(address);

        real_kv.call("testTimeOut");
    },

    testOom: function(address) {
        var real_kv = new Blockchain.Contract(address);
        real_kv.call("testOom");
    },

    testTpsForNormalCall: function() {
        
    },

    testTpsForMutiNvm: function(address) {
        console.log(address);
        var contractInterface = {
            testTps: function(){

            },
        };
        var v = Blockchain.transaction.value;
        var real_kv = new Blockchain.Contract(address);
        return real_kv.value(1).call("testTps");
    },

    save: function(address, key, value) {

        var real_kv  = new Blockchain.Contract(address);

        real_kv.value(2000000000000000000).call("save", key, value);
    },

    testFuncNotExist: function(address, key, value) {
        var real_kv  = new Blockchain.Contract(address);

        real_kv.call("testFuncNotExist", key, value);
    },

    testUsageOfValue: function(address, key, value) {
        var real_kv = new Blockchain.Contract(address);  
        real_kv.value(2000000000000000000);
        real_kv.call("saveWithNoValue", key, value);
    },

    saveByCall: function(address, key, value) {

        var real_kv  = new Blockchain.Contract(address);
        
        real_kv.value(2000000000000000000).call('save', key, value);
    },

    safeSave : function(address, key, value) {

        var real_kv  = new Blockchain.Contract(address);
    
        try {
            real_kv.value(2000000000000000000).call("save", key, value);
        } catch(err) {
            var value = Blockchain.transaction.value;
            real_kv.value(value).call("save", key, value);
        }
    },

    get: function(address, key) {
        var real_kv = new Blockchain.Contract(address);
        var args = new Array();
        args[0] = key;
        return real_kv.call("get", key)
    },

    testTryCatch: function(address) {
        var real_kv = new Blockchain.Contract(address);

        try {
            real_kv.call("throwErr");
        } catch(err) {
            return;
        }
    },

    accept: function() {

    },
}

module.exports = kvStore;
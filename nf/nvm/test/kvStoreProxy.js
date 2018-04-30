"use strict"

var kvStore = function() {
};

var contractInterface = {
    save: function() {
    },
    get: function() {
    }
}


kvStore.prototype = {
    init : function() {
        //
    },

    save : function(address, key, value) {

        var real_kv  = new Blockchain.Contract(address, contractInterface);
        
        // if (Blockchain.transaction.value < 2) {
        //     throw("nas is not enough")
        // }

        var args = new Array();
        args[0] = key;
        args[1] = value;
        // real_kv.value(2000000000000000000).call('save', JSON.stringify(args));
        real_kv.value(2000000000000000000).save(key, value);
    },

    get: function(address, key) {
        var real_kv = new Blockchain.Contract(address, contractInterface);
        var args = new Array();
        args[0] = key;
        return real_kv.get(key)
        //return real_kv.call('get', JSON.stringify(args));
    }

}

module.exports = kvStore;
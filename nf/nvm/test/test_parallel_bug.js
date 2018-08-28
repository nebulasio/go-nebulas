"use strict";

var parallelTest = function () {
    LocalContractStorage.defineProperties(this, {
        val: null, 
    });
};

parallelTest.prototype = {
    init: function (a) {
        this.val = a;
    },

     test: function() {
        var a = this.val;
        if (!a) {
             throw("reproduce the bug")
        }
        this.val = a + 1
        return this.val
     }
};
module.exports = parallelTest;
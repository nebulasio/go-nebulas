// since version 1.0.5

'use strict';

var crypto = require('crypto.js');


var Contract = function() {

};

Contract.prototype = {
    init: function(){},

    testSha256: function(data) {
        return crypto.sha256(data);
    },

    testSha3256: function(data) {
        return crypto.sha3256(data);
    },

    testRipemd160: function(data) {
        return crypto.ripemd160(data);
    },

    testRecoverAddress: function(alg, hash, sign) {
        return crypto.recoverAddress(alg, hash, sign);
    },

    testMd5: function(data) {
        return crypto.md5(data);
    },

    testBase64: function(data) {
        return crypto.base64(data);
    }
};

module.exports = Contract;

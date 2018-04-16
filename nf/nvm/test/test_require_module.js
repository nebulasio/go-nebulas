'use strict';

var SampleContract = function () {
    LocalContractStorage.defineProperties(this, {
    });
};

SampleContract.prototype = {
    init: function () {
    },
    requireNULL: function () {
        require("\x00");
    },
    requireNotExistPath: function () {
        _native_require("../../../../../../../../etc/");
    },
    requireCurPath: function () {
        _native_require("lib");
    },
    requireNotExistFile: function () {
        _native_require("lib.js");
    },
    
};

module.exports = SampleContract;

'use strict';

var Mixin = function () {};

Mixin.UNPAYABLE = function () {
    if (Blockchain.transaction.value.gt(0)) {
        return false;
    }
    return true;
};

Mixin.PAYABLE = function () {
    if (Blockchain.transaction.value.gt(0)) {
        return true;
    }
    return false;
};

Mixin.POSITIVE = function () {
    console.log("POSITIVE");
    return true;
};

Mixin.UNPOSITIVE = function () {
    console.log("UNPOSITIVE");
    return false;
};

Mixin.decorator = function () {
    var funcs = arguments;
    if (funcs.length < 1) {
        throw new Error("mixin decorator need parameters");
    }

    return function () {
        for (var i = 0; i < funcs.length - 1; i ++) {
            var func = funcs[i];
            if (typeof func !== "function" || !func()) {
                throw new Error("mixin decorator failure");
            }
        }

        var exeFunc = funcs[funcs.length - 1];
        if (typeof exeFunc === "function") {
            exeFunc.apply(this, arguments);
        } else {
            throw new Error("mixin decorator need an executable method");
        }
    };
};

var SampleContract = function () {
};

SampleContract.prototype = {
    init: function () {
    },
    unpayable: function () {
        console.log("contract function unpayable:", arguments);
    },
    payable: Mixin.decorator(Mixin.PAYABLE, function () {
        console.log("contract function payable:",arguments);
    }),
    contract1: Mixin.decorator(Mixin.POSITIVE, function (arg) {
        console.log("contract1 function:", arg);
    }),
    contract2: Mixin.decorator(Mixin.UNPOSITIVE, function (arg) {
        console.log("contract2 function:", arg);
    }),
    contract3: Mixin.decorator(Mixin.PAYABLE, Mixin.POSITIVE, function (arg) {
        console.log("contract3 function:", arg);
    }),
    contract4: Mixin.decorator(Mixin.PAYABLE, Mixin.UNPOSITIVE, function (arg) {
        console.log("contract4 function:", arg);
    })
};

module.exports = SampleContract;

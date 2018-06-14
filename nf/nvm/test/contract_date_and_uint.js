// since version 1.0.5

"use strict";

var Uint64 = Uint.Uint64;
var Uint128 = Uint.Uint128;
var Uint256 = Uint.Uint256;
var Uint512 = Uint.Uint512;

var Contract = function() {

};

Contract.prototype = {
    init: function(){},

    testDate: function() {

        Event.Trigger("testDate.arguments", {
			args: arguments
        });

        var date = arguments.length == 0 ? new Date() : new Date(arguments[0]);
        var date2 = new Date();
        date2.setFullYear(1988);
        var data = {
            UTC: Date.UTC(),
            now: Date.now(),
            parse: Date.parse('04 Dec 1995 00:12:00 GMT'),
            getUTCDate: date.getUTCDate(),
            toJSON: date.toJSON(),
            setFullYear: date2.toString(),
            height: Blockchain.block.height,
            timestamp: Blockchain.block.timestamp,
            valueOf: date.valueOf(),
            date_toString: date.toString(),
            date_getTime: date.getTime(),
            date_getFullYear: date.getFullYear(),
            equalBlockTime: Blockchain.block.timestamp == (date.getTime() / 1000),
            toLocaleString: date.toLocaleString(),
            toLocaleDateString: date.toLocaleDateString(),
            toLocaleTimeString: date.toLocaleTimeString(),
            getTimezoneOffset: date.getTimezoneOffset()
        };

        Event.Trigger("Date", {
			data: data
        });

        return data;
    },

    testUint64: function() {
        var n1 = new Uint64("10000");
        var n2 = new Uint64("2000000");
        var n3 = new Uint512("2000000");

        var overflow = false;
        try {
            n2.pow(n1);
        } catch (err) {
            if (err.message === "[Uint64 Error] overflow") {
                overflow = true;
            }
        }
        var underflow = false;
        try {
            n1.minus(n2);
        } catch (err) {
            if (err.message === "[Uint64 Error] underflow") {
                underflow = true;
            }
        }

        var isNaN = false;
        try {
            n1.div(new Uint64(0));
        } catch (err) {
            if (err.message === "[Uint64 Error] not an integer") {
                isNaN = true;
            }
        }

        var isNaN2 = false;
        try {
            new Uint64(1.2);
        } catch (err) {
            if (err.message === "[Uint64 Error] not an integer") {
                isNaN2 = true;
            }
        }

        var incompatible = false;
        try {
            n1.plus(n3);
        } catch (err) {
            if (err.message === "[Uint64 Error] incompatible type") {
                incompatible = true;
            }
        }

        var data = {
            "n1": n1.toString(10),
            "n2": n2.toString(10),
            "n2pown1Overflow": overflow, 
            "n1minusn2Underflow": underflow,
            "plus": n1.plus(n2).toString(10),
            "minus": n2.minus(n1).toString(10),
            "mul": n1.mul(n2).toString(10),
            "div": n2.div(n1).toString(10),
            "mod": n2.mod(n1).toString(10),
            "pow": n1.pow(new Uint64(2)).toString(10),
            "n2gtn1": n2.cmp(n1) > 0,
            "n1isZero": n1.isZero(),
            "0isZero": new Uint64(0).isZero(),
            "n1div0NaN": isNaN,
            "floatNaN": isNaN2,
            "incompatible": incompatible
        };

        Event.Trigger("Uint64", {
			data: data
        });
        return data;
    }
};

module.exports = Contract;
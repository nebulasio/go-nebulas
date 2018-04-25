"use strict";

var Contract = function() {

};

Contract.prototype = {
    init: function(){},
    accept: function(){},

    testDate: function() {

        Event.Trigger("testDate.arguments", {
			args: arguments
        });

        var date = arguments.length == 0 ? new Date() : new Date(arguments[0]);
        var data = {
            now: Date.now(),
            height: Blockchain.block.height,
            timestamp: Blockchain.block.timestamp,
            valueOf: date.valueOf(),
            date_toString: date.toString(),
            date_getTime: date.getTime(),
            date_getFullYear: date.getFullYear(),
            equalBlockTime: Blockchain.block.timestamp == (date.getTime() / 1000)
        };
        Event.Trigger("Date", {
			data: data
        });

        date.setFullYear(2000);
        Event.Trigger("Date.modi", {
			data: {
                valueOf: date.valueOf(),
                date_toString: date.toString(),
                date_getTime: date.getTime(),
                date_getFullYear: date.getFullYear(),
                equalBlockTime: Blockchain.block.timestamp == (date.getTime() / 1000)
            }
        });
        
        return data;
    },

    testRandom: function() {
        // TODO
    }
};

module.exports = Contract;
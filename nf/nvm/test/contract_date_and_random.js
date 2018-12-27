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

    testDate2: function() {
        var date = new Date();
        
        Event.Trigger("Date.TZ", {
            "timezone": date.getTimezoneOffset(),
            "toLocaleString": date.toLocaleString(),
            "toLocaleDateString": date.toLocaleDateString(),
            "toLocaleTimeString": date.toLocaleTimeString()
        });
    },

    testRandom: function(userseed) {
        var r1 = Math.random();
        var r12 = Math.random();
        var r13 = Math.random();
        // equivalent to reset seed
        Math.random.seed(userseed);
        var r2 = Math.random();

        Event.Trigger("random", {
            "seed": Blockchain.block.seed, 
            "defaultSeedRandom1": r1,
            "defaultSeedRandom2": r12,
            "defaultSeedRandom3": r13,
            "userSeedRandom": r2
        });
    },

    testRandom1_1_0: function(userseed) {
        var r1 = Math.random();
        var r12 = Math.random();
        var r13 = Math.random();
        var r2 = Math.random();

        Event.Trigger("random", {
            "seed": Blockchain.block.seed, 
            "defaultSeedRandom1": r1,
            "defaultSeedRandom2": r12,
            "defaultSeedRandom3": r13,
            "userSeedRandom": r2
        });
    }
};

module.exports = Contract;
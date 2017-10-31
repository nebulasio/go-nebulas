'use strict';

function Console() {}

function log(...args) {
    var level = args.shift();
    if (typeof (level) != 'number') {
        throw 'level must be number.';
    }

    var msg = '';
    for (var i = 0; i < args.length - 1; i++) {
        msg += format(args[i]) + ' ';
    }
    msg += format(args[args.length - 1]);

    _native_log(level, msg);
}

function format(obj) {
    if (typeof (obj) == 'object') {
        return JSON.stringify(obj);
    }
    return obj;
}

[
    ['debug', 1],
    ['warn', 2],
    ['info', 3],
    ['log', 3],
    ['error', 4]
].forEach(function (val) {
    Console.prototype[val[0]] = log.bind(null, val[1]);
});

module.exports = new Console();
module.exports.Console = Console;

// test 1.
_native_log(1, "log from test_console.js.");

// test 2.
// var console = require('console.js');
console.info('log data:', 1, true, undefined, {
    'x': 1
}, console.error);

// test 3.
var _test_3_z = {};
console.log('typeof(this) == ', typeof (this), this);

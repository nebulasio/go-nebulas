#!/usr/bin/env node

const fs = require('fs');
const instCounter = require('../lib/instruction_counter.js');

function help() {
    console.log('JS file missing.');
};

if (process.argv.length < 3) {
    help();
    return;
}

var source_file = process.argv[2];
fs.readFile(source_file, (err, data) => {
    if (err) throw err;
    var source = data.toString();
    var ret = instCounter.processScript(source);
    console.log(ret.traceableSource);
});

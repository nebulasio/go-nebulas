'use strict';

var schedule = require('node-schedule');
var fs = require("fs");
var exec = require('child_process').exec;

// send transactions per 10 minute.
var j = schedule.scheduleJob('*/10 * * * *', function(){
    console.log("start transaction test");
    sendTransactionsTest();
});

var type = 0;

function sendTransactionsTest() {

    // var type = Math.floor(Math.random()*4);
    switch(type)  {
        case 0:
            startMochaTest("binary/value.test.js");
        case 1:
            startMochaTest("contract/contract.deploy.test.js");
        case 2:
            startMochaTest("contract/contract.call.test.js");
        case 3:
            startMochaTest("contract/contract.bankvault.test.js");
        case 4:
            startMochaTest("contract/contract.nrc20.test.js");
    }

    type++;
    type = type%5;
}


function startMochaTest(file) {
    const filePath = "/neb/app/logs/transactionTestResult.txt";
    // const filePath = "testResult.txt";

    var cmd = "mocha cases/" + file + " testneb1 -t 200000";
    console.log("start mocha:", cmd);
    exec(cmd, function (err, stdout, stderr) {
        var run = "cmd:" + cmd;
        var runResult = "run:" + (err ? false : true);
        var content = run + "\n" + runResult + "\n" +  "test result:\n" + stdout + "\n";
        content += stderr + "\n";

        fs.appendFile(filePath, content, 'utf8', function(err){
            if(err)
            {
                console.log(err);
            }
        });

    });
}

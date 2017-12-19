var fs = require('fs');
var net = require('net');
var shell = require('shelljs');

var ConfFolder = "conf/";
var BlockConfFolder = ConfFolder + "block/";
var TxConfFolder = ConfFolder + "transaction/";
var DownloadConfFolder = ConfFolder + "download/";

var client = new net.Socket();

client.connect(51413, '127.0.0.1', function () {
    console.log('Connected');
    tests(BlockConfFolder, sendBlock);
    tests(TxConfFolder, sendTx);
    tests(DownloadConfFolder, downloadBlock);
});

client.on('data', function (data) {
    console.log('Received: ' + data);
});

client.on('close', function () {
    console.log('Connection closed');
});

function sendBlock(conf) {
    shell.exec('./neb ...' + conf, function (code, stdout, stderr) {
        var block = new Buffer(stdout, 'base64');
        console.log(block.toString());
        client.write(block);
    })
}

function sendTx(conf) {
    shell.exec('./neb ...' + conf, function (code, stdout, stderr) {
        var block = new Buffer(stdout, 'base64');
        console.log(block.toString());
        client.write(block);
    })
}

function downloadBlock(conf) {
    shell.exec('./neb ...' + conf, function (code, stdout, stderr) {
        var block = new Buffer(stdout, 'base64');
        console.log(block.toString());
        client.write(block);
    })
}

function tests(folder, func) {
    fs.readdir(folder, (err, files) => {
        files.forEach(file => {
            console.log(file);
            func(file);
        });
    })
}

'use strict';

var Neblet = require('./neblet');
var os = require('os');
var expect = require('chai').expect;

var ipArray = ['127.0.0.1'];
var jsAgentArray = new Array();

for (var i=0;i<ipArray.length;i++) {
    var server = new Neblet(ipArray[i], 10000, 8090);
    var jsAgent = server.NebJs();
    jsAgentArray.push(jsAgent);
}

var coinbase;
var addresses;
var nonce = 0;
var first = true;

var intervaltime = 1000;
var myFunction = function(){
    clearInterval(interval);
    sendTransaction();
    intervaltime = Math.floor(Math.random()*1000);
    console.log('interval time: ' + intervaltime);
    interval = setInterval(myFunction, intervaltime);
};
var interval = setInterval(myFunction, intervaltime);


Array.prototype.indexOf = function(val) {
    for (var i = 0; i < this.length; i++) {
        if (this[i] == val) return i;
    }
    return -1;
};

Array.prototype.remove = function(val) {
    var index = this.indexOf(val);
    if (index > -1) {
    this.splice(index, 1);
    }
};


function sendTransaction() {
    var random = Math.floor(Math.random()*ipArray.length);
    var jsAgent = jsAgentArray[random];

    // get coinbase
    var nebState = jsAgent.api.getNebState();
    coinbase = nebState.coinbase;

    // get accounts
    var accounts = jsAgent.api.accounts();
    addresses = accounts.addresses;
    addresses.remove(coinbase);

    // unlock coinbase account
    var result = jsAgent.admin.unlockAccount(coinbase, 'zaq12wsx');
    if (!result.result) {
        console.error('unlock account error.');
        // clearInterval(test);
    }

    // get coinbase account state
    var accountState = jsAgent.api.getAccountState(coinbase);
    if (first) {
        nonce = parseInt(accountState.nonce);
        first = false;
    }
    var randomAddressFlag = Math.floor(Math.random()*addresses.length);

    // send transactions
    var txhash = jsAgent.api.sendTransaction(coinbase, addresses[randomAddressFlag], 10, nonce+1);
    console.log(txhash);
    if (txhash.code == 2) {
        if (txhash.error == 'key not unlocked') {
            jsAgent.admin.unlockAccount(coinbase, 'zaq12wsx');
        }
    } else {
        nonce++;
    }
}





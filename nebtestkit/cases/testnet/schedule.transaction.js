'use strict';

var schedule = require('node-schedule');
var FS = require("fs");

var Neb = require("../../../cmd/console/neb.js/lib/wallet").Neb;
var neb = new Neb();
neb.setRequest(new Neb.HttpRequest("https://testnet.nebulas.io"));

var Account = require("../../../cmd/console/neb.js/lib/wallet").Account;
var Transaction = require("../../../cmd/console/neb.js/lib/wallet").Transaction;
var Utils = require("../../../cmd/console/neb.js/lib/wallet").Utils;

var txCount = 8;
var TestnetChainID = 1001;

// send transactions per 2 minute.
var j = schedule.scheduleJob('*/2 * * * *', function(){
    console.log("start send transaction");
    sendTransactions();
});

function sendTransactions() {
    var from = Account.NewAccount();
    var email = Math.random() + "test@demo.io";
    var api = "/claim/api/claim/"+ email+ "/"+ from.getAddressString();
    neb.api.request("GET", api, null).then(function (resp) {
        console.log("claim tokens:" + JSON.stringify(resp));
    });
    // sleep to wait package claim transaction.
    sleep(6000);

    neb.api.getAccountState(from.getAddressString()).then(function (resp) {
        var balance = Utils.toBigNumber(resp.balance);
        for (var i = 0; i < txCount; i++) {
            var type = Math.floor(Math.random()*8);
            var nonce = parseInt(resp.nonce) + i + 1;
            switch(type)  {
                case 0:
                    sendNormal(from, balance, nonce);
                case 1:
                    sendGasLimitTnsufficient(from, balance, nonce);
                case 2:
                    sendBalanceInsufficient(from, balance, nonce);
                case 3:
                    sendFromToEqual(from, balance, nonce);
                case 4:
                    sendNoncebelow(from, balance, nonce);
                case 5:
                    sendNonceHeigh(from, balance, nonce);
                case 6:
                    sendContract(from, balance, nonce);
                case 7:
                    sendContractGasLimitTnsufficient(from, balance, nonce);
            }
        }
    });
}

function sendNormal(from, balance, nonce) {

    var to = Account.NewAccount();
    var tx = new Transaction(TestnetChainID, from, to, balance.divToInt(txCount), nonce);
    tx.signTransaction();
    neb.api.sendRawTransaction(tx.toProtoString());
}

function sendGasLimitTnsufficient(from, balance, nonce) {
    var to = Account.NewAccount();
    var tx = new Transaction(TestnetChainID, from, to, balance.divToInt(txCount), nonce, "0", "1");
    tx.signTransaction();
    neb.api.sendRawTransaction(tx.toProtoString());
}

function sendBalanceInsufficient(from, balance, nonce) {
    var to = Account.NewAccount();
    var tx = new Transaction(TestnetChainID, from, to, balance, nonce);
    tx.signTransaction();
    neb.api.sendRawTransaction(tx.toProtoString());
}

function sendFromToEqual(from, balance, nonce) {
    var tx = new Transaction(TestnetChainID, from, from, balance.divToInt(txCount), nonce);
    tx.signTransaction();
    neb.api.sendRawTransaction(tx.toProtoString());
}

function sendNoncebelow(from, balance, nonce) {
    var to = Account.NewAccount();
    var tx = new Transaction(TestnetChainID, from, to, balance.divToInt(txCount), nonce-1);
    tx.signTransaction();
    neb.api.sendRawTransaction(tx.toProtoString());
}

function sendNonceHeigh(from, balance, nonce) {
    var to = Account.NewAccount();
    var tx = new Transaction(TestnetChainID, from, to, balance.divToInt(txCount), nonce+2);
    tx.signTransaction();
    neb.api.sendRawTransaction(tx.toProtoString());
}

function sendContract(from, balance, nonce) {
    var to = Account.NewAccount();
    var bank = FS.readFileSync("../../../nf/nvm/test/bank_vault_contract.js", "utf-8");
    var contract = {
        "source": bank,
        "sourceType": "js",
        "args": ""
    };
    var value = balance.divToInt(txCount);
    neb.api.estimateGas(from, to, value, nonce, "0", "0",contract).then(function (resp) {
        var tx = new Transaction(TestnetChainID, from, to, value, nonce, "1000000", resp.estimate_gas, contract);
        tx.signTransaction();
        neb.api.sendRawTransaction(tx.toProtoString());
    });
}

function sendContractGasLimitTnsufficient(from, balance, nonce) {
    var to = Account.NewAccount();
    var bank = FS.readFileSync("../../../nf/nvm/test/bank_vault_contract.js", "utf-8");
    var contract = {
        "source": bank,
        "sourceType": "js",
        "args": ""
    };
    var value = balance.divToInt(txCount);
    neb.api.estimateGas(from, to, value, nonce, "0", "0",contract).then(function (resp) {
        var gasLimit = Utils.toBigNumber(resp.estimate_gas).sub("10");
        var tx = new Transaction(TestnetChainID, from, to, value, nonce, "1000000", gasLimit, contract);
        tx.signTransaction();
        neb.api.sendRawTransaction(tx.toProtoString());
    });
}


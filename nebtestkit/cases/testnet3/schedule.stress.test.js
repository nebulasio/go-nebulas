'use strict';
var schedule = require('node-schedule');
var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var HttpRequest = require("../../node-request");
var FS = require("fs");
var logStream = FS.createWriteStream('stress.test.log', {flags: 'a'});
process.stdout.write = process.stderr.write = logStream.write.bind(logStream)

var env; // local testneb1 testneb2
var AddressNumber;
var SendTimes;

var args = process.argv.splice(2);

if (args.length !=3 ){
    // give default config
    env = "testneb3";
    AddressNumber = 200;
    SendTimes = 30;
} else {
    env = args[0]; // local testneb1 testneb2

    AddressNumber = parseInt(args[1]);
    SendTimes = parseInt(args[2]);
}

if (AddressNumber <=0 || SendTimes <=0 ){

    console.log("please input correct AddressNumber and SendTimes");
    return;
}

var Neb = Wallet.Neb;
var neb = new Neb();

var ChainID;
var from;
var accountArray;

//local
if (env == 'local'){
    neb.setRequest(new HttpRequest("http://127.0.0.1:8685"));//https://testnet.nebulas.io
    ChainID = 100;
    from = new Wallet.Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
}else if(env == 'testneb1'){
    neb.setRequest(new HttpRequest("http://13.57.245.249:8685"));
    ChainID = 1001;
    from = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
}else if(env == "testneb2"){
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
    ChainID = 1002;
    from = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
}else if(env == "testneb3"){
    neb.setRequest(new HttpRequest("http://35.177.214.138:8685"));
    ChainID = 1003;
    from = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
}else{
    console.log("please input correct env local testneb1 testneb2");
    return;
}

var lastnonce = 0;

var j = schedule.scheduleJob('10,40 * * * *', function(){
    console.log("start transaction stress test");
    neb.api.getAccountState(from.getAddressString()).then(function (resp) {

        console.log("master accountState resp:" + JSON.stringify(resp));
        lastnonce = parseInt(resp.nonce);
        console.log("lastnonce:", lastnonce);

        sendTransactions();
    });
});

function sendTransactions() {

    accountArray = new Array();
    for (var i = 0; i < AddressNumber; i++) {
        var account = Wallet.Account.NewAccount();
        accountArray.push(account);
    }

    var type = Math.floor(Math.random()*2);
    switch(type)  {
        case 0:
            console.log("send normal transactions!!!");
            sendNormalTransactions(SendTimes, "1");
            break;
        case 1:
            console.log("send contract transactions!!!");
            sendContractTransactions();
            break;
    }
}

function sendNormalTransactions(totalTimes, value) {
    var nonce = lastnonce;
    for (var i = 0; i < AddressNumber; i++) {
        sendTransaction(0, totalTimes, value, nonce, accountArray[i]);
        nonce = nonce + totalTimes;
    }
}

function sendTransaction(sendtimes, totalTimes, value, nonce, address) {
    if (sendtimes < totalTimes) {
        var transaction = new Wallet.Transaction(ChainID, from, address, value, ++nonce);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();
        neb.api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw transaction resp:" + JSON.stringify(resp));
            sendtimes++;
            if (resp.txhash) {
                sendTransaction(sendtimes, totalTimes, value, nonce, address);
            }
        });
    }

}

function sendContractTransactions() {
    //claim tokens
    sendNormalTransactions(1, neb.nasToBasic(1));
    lastnonce += AddressNumber;

    var intervalAccount = setInterval(function () {
        neb.api.getAccountState(from.getAddressString()).then(function (resp) {
            // console.log("master accountState resp:" + JSON.stringify(resp));
            var nonce = parseInt(resp.nonce);
            console.log("lastnonce:", lastnonce, "resp_nonce:", nonce);

            if (lastnonce <= nonce){
                clearInterval(intervalAccount);
                deployContract();
            }
        });
    }, 2000);
}

function deployContract(){

    var nonce = lastnonce;
    console.log("deploy contract");

    neb.api.getAccountState(from.getAddressString()).then(function (state) {
        lastnonce = parseInt(state.nonce);
        console.log("lastnonce:", lastnonce);

        // create contract
        var bank = FS.readFileSync("/neb/golang/src/github.com/nebulasio/go-nebulas/nf/nvm/test/bank_vault_contract.js", "utf-8");
        var contract = {
            "source": bank,
            "sourceType": "js",
            "args": ""
        };

        var transaction = new Wallet.Transaction(ChainID, from, from, "0", ++lastnonce, "1000000", "2000000", contract);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();
        neb.api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw contract transaction resp:" + JSON.stringify(resp));
            // ContractHash = resp.txhash;
            // ContractAddress = resp.contract_address;

            checkContractDeployed(resp.txhash);
        });
    });
}

function checkContractDeployed(ContractHash){

    var retry = 0;

    // contract status and get contract_address
    var interval = setInterval(function () {
        console.log("getTransactionReceipt hash:"+ContractHash);
        neb.api.getTransactionReceipt(ContractHash).then(function (resp) {

            console.log("tx receipt:" + resp.status);

            if(resp.status && resp.status === 1) {
                clearInterval(interval);
                sendMutilContractTransaction(resp.contract_address);
            }
        }).catch(function (err) {
            retry++;
            console.log("error!", JSON.stringify(err.error));
            if (retry > 10) {
                console.log("deploy contract failed");
                console.log(JSON.stringify(err.error));
                clearInterval(interval);
            }
        });

    }, 2000);
}

function sendMutilContractTransaction(contract){
    for (var i = 0; i < AddressNumber; i++) {
        sendContractTransaction(0, 0, accountArray[i], contract);
    }
}

function sendContractTransaction(sendtimes, nonce, from_address, contract_address) {
    if(sendtimes < SendTimes) {
        var call = {
            "function": "save",
            "args":"[10000]"
        };

        console.log("send contract nonce:",nonce);
        var transaction = new Wallet.Transaction(ChainID, from_address, contract_address, "0", ++nonce, "1000000", "2000000", call);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();
        neb.api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw contract transaction resp:" + JSON.stringify(resp));
            sendtimes++;
            if(resp.txhash) {
                sendContractTransaction(sendtimes, nonce, from_address, contract_address);
            }
        });
    }

}
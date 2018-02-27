'use strict';

var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var sleep = require("system-sleep");
var HttpRequest = require("../../node-request");
var FS = require("fs");

var args = process.argv.splice(2);

if (args.length !=3 ){
	console.log("please input args 0:env(local,testneb1,testneb2) 1:address number(concurrency) 2:sendtimes")
	return;
}

var env = args[0]; // local testneb1 testneb2

const AddressNumber = parseInt(args[1]);
const SendTimes = parseInt(args[2]);

if (AddressNumber <=0 || SendTimes <=0 ){
	
	console.log("please input correct AddressNumber and SendTimes");
	return;
}

var Neb = Wallet.Neb;
var neb = new Neb();

var ChainID;
var source, deploy;

//local
if (env == 'local'){
	neb.setRequest(new HttpRequest("http://127.0.0.1:8685"));//https://testnet.nebulas.io
	ChainID = 100;
    source = new Wallet.Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
}else if(env == 'testneb1'){
	neb.setRequest(new HttpRequest("http://35.182.48.19:8685"));
	ChainID = 1001;
    source = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
}else if(env == "testneb2"){
	neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
	ChainID = 1002;
    source = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
}else{
	console.log("please input correct env local testneb1 testneb2")
	return;
}

var lastnonce = 0;

// new account  to get address
var accountArray = new Array();
for (var i = 0; i < AddressNumber; i++) {
    var account = Wallet.Account.NewAccount();
    //var hash = account.getAddressString();
    accountArray.push(account);
}

neb.api.getAccountState(source.getAddressString()).then(function (resp) {
    console.log("source account state:" + JSON.stringify(resp));
    lastnonce = parseInt(resp.nonce);

    deploy = Wallet.Account.NewAccount();

    var accounts = new Array();
    accounts.push(deploy);
    cliamTokens(accounts, neb.nasToBasic(10), function () {
    	deployContract();
	});

});

function cliamTokens(accounts, value, done) {
    for (var i = 0; i < accounts.length; i++) {
        sendTransaction(source, accounts[i], value, ++lastnonce);
        sleep(30);
    }
    checkCliamTokens(done);
}

function checkCliamTokens(done) {
    var intervalAccount = setInterval(function () {
        neb.api.getAccountState(source.getAddressString()).then(function (resp) {
            // console.log("master accountState resp:" + JSON.stringify(resp));
            var nonce = parseInt(resp.nonce);
            console.log("check cliam tokens lastnonce:", lastnonce);

            if (lastnonce <= nonce){
                clearInterval(intervalAccount);

                done();
            }
        });
    }, 2000);
}

function deployContract(){

    // create contract
    var source = FS.readFileSync("../nf/nvm/test/NRC20.js", "utf-8");
    var contract = {
        "source": source,
        "sourceType": "js",
        "args": "[\"StandardToken\", \"NRC\", 18, \"1000000000\"]"
    };

    var transaction = new Wallet.Transaction(ChainID, deploy, deploy, "0", 1, "0", "20000000000", contract);
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();

    // console.log("contract:" + rawTx);
    
    neb.api.sendRawTransaction(rawTx).then(function (resp) {
        console.log("send raw contract transaction resp:" + JSON.stringify(resp));

        checkContractDeployed(resp.txhash);
    });
}

function checkContractDeployed(txhash){

    var retry = 0;

    // contract status and get contract_address 
    var interval = setInterval(function () {
		// console.log("getTransactionReceipt hash:"+txhash);
		neb.api.getTransactionReceipt(txhash).then(function (resp) {

			console.log("deploy status:" + resp.status);
			
			if(resp.status && resp.status === 1) {
                clearInterval(interval);

                console.log("deploy private key:" + deploy.getPrivateKeyString());
                console.log("deploy address:" + deploy.getAddressString());
                console.log("deploy contract address:" + resp.contract_address);
                // console.log("deploy receipt:" + JSON.stringify(resp));

                checkNRCBalance(deploy.getAddressString(), resp.contract_address);

                sendContractTransactions(resp.contract_address);
			}
		}).catch(function (err) {
			retry++;
			console.log("retry:", retry);
			if (retry > 10) {
				console.log(JSON.stringify(err.error));
				clearInterval(interval);
			}
		});

    }, 2000);
}


function sendTransaction(from, address, value, nonce) {
    var transaction = new Wallet.Transaction(ChainID, from, address, value, nonce);
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();
    neb.api.sendRawTransaction(rawTx).then(function (resp) {
        console.log("send raw transaction resp:" + JSON.stringify(resp));
    });
}

function checkNRCBalance(address, contractAddress) {
    var contract = {
        "function": "balanceOf",
        "args": "[\"" + address + "\"]"
    };

    neb.api.call(address, contractAddress, "0", 2, "0", "0", contract).then(function (resp) {
        console.log("balance of NRC:" + JSON.stringify(resp));
    });
}

function sendContractTransactions(contract) {
	neb.api.getAccountState(deploy.getAddressString()).then(function (resp) {

		console.log("contract state:", JSON.stringify(resp));

		var nonce = parseInt(resp.nonce);

		for (var i = 0; i < accountArray.length; i++) {
            sendContractTransaction(0, deploy, accountArray[i].getAddressString(), nonce, contract);
		}

		sleep(10000);

        for (var i = 0; i < accountArray.length; i++) {
        	checkNRCBalance(accountArray[i].getAddressString(), contract);
        }
    });
}

function sendContractTransaction(sendtimes, from, to, nonce, contract) {
    if(sendtimes < SendTimes) {
        var call = {
            "function": "transfer",
            "args": "[\"" + to + "\", "+ "10" + "]"
        };

		console.log("send contract nonce:",nonce);
        var transaction = new Wallet.Transaction(ChainID, from, contract, "0", ++nonce, "0", "2000000000", call);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();
        neb.api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw contract transaction resp:" + JSON.stringify(resp));
            sendtimes++;
            if(resp.txhash) {
                sendContractTransaction(sendtimes, from, to, nonce, contract);
            }
        });
    } 
    
}
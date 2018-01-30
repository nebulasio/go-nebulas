'use strict';

var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var sleep = require("system-sleep");
var HttpRequest = require("../../node-request");

var Neb = Wallet.Neb;
var neb = new Neb();
//http://34.205.26.12:8685
neb.setRequest(new HttpRequest("http://127.0.0.1:8685"));//https://testnet.nebulas.io

var FS = require("fs");

const AddressNumber = 10;
const SendTimes = 10;
var lastnonce = 0;

var ChainID = 100;

// var master = Wallet.Account.NewAccount();
// var from = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
var from = new Wallet.Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
// 1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c

// new account  to get address
var accountArray = new Array();
for (var i = 0; i < AddressNumber; i++) {
    var account = Wallet.Account.NewAccount();
    //var hash = account.getAddressString();
    accountArray.push(account);
}

neb.api.getAccountState(from.getAddressString()).then(function (resp) {
    console.log("master accountState resp:" + JSON.stringify(resp));
    lastnonce = parseInt(resp.nonce);
    // console.log("lastnonce:", lastnonce);
});

sleep(2000);
var nonce = lastnonce;

var ContractHash ;
var ContractAddress;

deployContract(nonce);

function deployContract(nonce){
    // create contract
    var bank = FS.readFileSync("../nf/nvm/test/bank_vault_contract.js", "utf-8");
    var contract = {
        "source": bank,
        "sourceType": "js",
        "args": ""
    };

    var transaction = new Wallet.Transaction(ChainID, from, from, "0", ++nonce, "0", "20000000000", contract);
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();

    // console.log("contract:" + rawTx);
    
    neb.api.sendRawTransaction(rawTx).then(function (resp) {
        console.log("send raw contract transaction resp:" + JSON.stringify(resp));
        ContractHash = resp.txhash;
        ContractAddress = resp.contract_address;

        checkContractDeployed();
    });

}

function checkContractDeployed(){

    var retry = 0;

    // contract status and get contract_address 
    var interval = setInterval(function () {
        if (ContractHash){
            console.log("getTransactionReceipt hash:"+ContractHash);
            neb.api.getTransactionReceipt(ContractHash).then(function (resp) {

                console.log("tx receipt:" + resp.status);
                
                if(resp.status === 1) {

                    clearInterval(interval);
                    cliamTokens();
                }
            }).catch(function (err) {
                retry++;
                if (retry > 3) {
                    console.log(JSON.stringify(err.error));
                    clearInterval(interval);
                }
            });
        }

    }, 2000);
}

function cliamTokens() {
    for (var j = 0; j < AddressNumber; j++) {
        sendTransaction(nonce, accountArray[j]);
        sleep(10);
    }
}

function sendTransaction(nonce, address) {
    var transaction = new Wallet.Transaction(ChainID, from, address, neb.nasToBasic(1), ++nonce);
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();
    neb.api.sendRawTransaction(rawTx).then(function (resp) {
        console.log("send raw transaction resp:" + JSON.stringify(resp));
    });
}


// get current height
// var BeginHeight ;
//
// neb.api.getNebState().then(function (resp) {
//
//     BeginHeight = resp.height;
//     console.log("get NebState resp:" + JSON.stringify(resp));
// });
//
// sleep(5000);



function getTransactionNumberByHeight(){

    neb.api.getNebState().then(function (resp) {
        console.log("====================") 
        var EndHeight = resp.height
        console.log("BeginHeight:"+BeginHeight+ " EndHeight:"+EndHeight)
        
        //console.log("send raw transaction resp:" + JSON.stringify(resp));

        var height = BeginHeight
        for(;height<=EndHeight;height++){
            neb.api.getBlockByHeight(height, false).then(function (resp) {
                if(resp.transactions){
                    //console.log("master accountState resp:" + JSON.stringify(resp));
                    console.log(resp.height, resp.transactions.length)
                    return;
                }else{
                    console.log(resp.height, 0)
                }
            });
            sleep(10)
        }
    });
}


function sendMutilTransaction(address){

    var nonce = lastnonce;
    var t1 = new Date().getTime();
    for (var j = 0; j < AddressNumber; j++) {
        sendContractTransaction(0, nonce, accountArray[j], address);
        nonce = nonce + SendTimes;
    }


    sleep(1000*SendTimes)
    getTransactionNumberByHeight();
}



function sendContractTransaction(sendtimes, nonce, from_address, contract_address) {
    if(sendtimes < SendTimes) {
        var call = {
            "function": "totalSupply",
            "args":""
        }

        var transaction = new Wallet.Transaction(ChainID, from_address, contract_address, "0", ++nonce, "0", "20000000000", call);
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

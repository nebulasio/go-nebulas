'use strict';

var TestnetNodes = require('../../testnet-nodes');
var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var sleep = require("system-sleep");
const https = require("https");
var HttpRequest = require("../../node-request");

var Neb = require("../../../cmd/console/neb.js/lib/wallet").Neb;


//var expect = require('chai').expect
var BigNumber = require('bignumber.js');
var FS = require("fs");

const AddressNumber = 10;
const SendTimes = 10;
var lastnonce = 0;

var ChainID = 1002

//var nodes = new TestnetNodes();
//nodes.Start();

var neb = new Neb();
neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
//nodes.RPC(0)

// var master = Wallet.Account.NewAccount();
var from = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");

neb.api.getAccountState(from.getAddressString()).then(function (resp) {
    console.log("master accountState resp:" + JSON.stringify(resp));
    lastnonce = parseInt(resp.nonce);
    console.log("lastnonce:", lastnonce);
});
sleep(2000)

// new account  to get address 
var accountArray = new Array();
for (var i = 0; i < AddressNumber; i++) {
    var account = Wallet.Account.NewAccount();
    //var hash = account.getAddressString();
    accountArray.push(account);
}

// send transaction 
var nonce = lastnonce;
var t1 = new Date().getTime();
for (var j = 0; j < AddressNumber; j++) {
    sendTransaction(nonce, accountArray[j]);
    ++nonce;
}

function sendTransaction(nonce, address) {
    var transaction = new Wallet.Transaction(ChainID, from, address, "100000000000", ++nonce);
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();
    neb.api.sendRawTransaction(rawTx).then(function (resp) {
        console.log("send raw transaction resp:" + JSON.stringify(resp));
    });
}


// get current height 
var BeginHeight ;

neb.api.getNebState().then(function (resp) {

    BeginHeight = resp.height
    console.log("get NebState resp:" + JSON.stringify(resp));
});

sleep(5000);



// get tail hash height 




var ContractHash ;

var ContractAddress;

createContract(nonce)

sleep(1000)

sendContract();


function createContract(nonce){
    // create contract
    var erc20 = FS.readFileSync("../nf/nvm/test/ERC20.js", "utf-8");
    //console.log("erc20:"+erc20);

    var contract = {
        "source": erc20,
        "sourceType": "js",
        "args": '["NebulasToken", "NAS", 1000000000]'
    }

    //console.log('nonce:'+nonce)

    var transaction = new Wallet.Transaction(ChainID, from, from, "1", ++nonce, "0", "20000000000", contract);
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();
    
    neb.api.sendRawTransaction(rawTx).then(function (resp) {
        console.log("send raw contract transaction resp:" + JSON.stringify(resp));
        ContractHash = resp.txhash
        ContractAddress = resp.contract_address
    });

}


function sendContract(){

    // contract status and get contract_address 
    var interval = setInterval(function () {
        if (ContractHash){
            console.log("getTransactionReceipt hash:"+ContractHash)
            neb.api.getTransactionReceipt(ContractHash).then(function (resp) {

                //console.log("tx receipt:" + JSON.stringify(resp));
                
                if(resp.status == 1) {

                    clearInterval(interval);
                    sendMutilTransaction(ContractAddress)
                }
            }).catch(function (err) {
                console.log(JSON.stringify(err.error));
                clearInterval(interval);
            });
        }

    }, 2000);
}



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

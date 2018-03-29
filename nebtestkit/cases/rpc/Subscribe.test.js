'use strict';

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet = require("nebulas");

var coinbase,
    client,
    sourceAccount,
    chain_id,
    server_address;


var env = process.env.NET || 'local';
if (env === 'testneb1') {
    chain_id = 1001;
    sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
    server_address = "35.182.48.19:8684";
  
  } else if (env === "testneb2") {
    chain_id = 1002;
    sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
    server_address = "34.205.26.12:8684";
  
  } else if (env === "testneb3") {
    chain_id = 1003;
    sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
    server_address = "35.177.214.138:8684";
  
  } else if (env === "testneb4") { //super node
    chain_id = 1004;
    sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
    coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
    server_address = "35.154.108.11:8684";
  } else if (env === "testneb4_normalnode"){
    chain_id = 1004;
    sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
    coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
    server_address = "18.197.107.228:8684";
  } else if (env === "local") {
    chain_id = 100;
    sourceAccount = new Wallet.Account("d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5");
    coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
    server_address = "127.0.0.1:8684";
  
  } else if (env === "maintest"){
    chain_id = 2;
    sourceAccount = new Wallet.Account("d2319a8a63b1abcb0cc6d4183198e5d7b264d271f97edf0c76cfdb1f2631848c");
    coinbase = "n1dZZnqKGEkb1LHYsZRei1CH6DunTio1j1q";
    server_address = "54.149.15.132:8684";
  } else {
    throw new Error("invalid env (" + env + ").");
  }

var all_topics = [
    "chain.pendingTransaction",
    "chain.sendTransaction",
    "chain.deployContract",
    "chain.callContract",
    "chain.contract",
    "chain.delegate",
    "chain.candidate",
    "chain.linkBlock",
    "chain.latestIrreversibleBlock",
    "chain.executeTxFailed",
    "chain.executeTxSuccess",
    "chain.transactionResult"
];

function testSubscribe(testInput, testExpect, done) {
    var counter = 0;
    var call;
    try {
        call = client.Subscribe(testInput);
        
        call.on('end', () => {
            console.log("server finish sending");
        });

        call.on('data', evt => {
            console.log("[event] " + JSON.stringify(evt) + "\n");
            try {
                expect(evt).to.have.property('topic');
                expect(evt).to.have.property('data');
                counter++;
                if (testExpect.justRecieveNum <= counter) {
                    console.log("recieved " + counter + " events, now stop.")
                    call.cancel();
                    done();
                }
            } catch (err) {
                call.cancel();
                done(err);
            }
        });

        call.on('status', status => {
            console.log("[status] " + JSON.stringify(status) + "\n");
        });

        setTimeout(() => {
            if (0 == counter) {
                console.log(new Error("no events got after " + testExpect.timeout + " ms"))
                call.cancel();
                done();
            }
        }, testExpect.timeout);

    } catch(err) {
        console.log("call failed:" + err.toString())
        if (testExpect.callFailed) {
            try {
                expect(err.toString()).to.have.string(testExpect.errorMsg);
                done();
            } catch(er) {
                done(err);
            }
        } else {
            done(err)
        }
    }
}

describe("rpc: Subscribe", () => {
    before(() => {
        client = rpc_client.new_client(server_address);
    });

    it('1. subsribe all', done => {

        var testInput = {
            topics: all_topics
        }

        var testExpect = {
            callFailed: false,
            justRecieveNum: 1,
            timeout: 20000
        }

        testSubscribe(testInput, testExpect, done)
    });

    it('2. all topics are unkown', done => {

        var testInput = {
            topics: ["unkown"]
        }

        var testExpect = {
            callFailed: false,
            justRecieveNum: 1,
            timeout: 20000
        }

        testSubscribe(testInput, testExpect, done)
    });

    it('3. partial of all topics are unkown', done => {

        var testInput = {
            topics: [
                "unkown",
                "chain.linkBlock"
            ]
        }

        var testExpect = {
            callFailed: false,
            justRecieveNum: 1,
            timeout: 20000
        }

        testSubscribe(testInput, testExpect, done)
    }); 

    it('4. no topic', done => {

        var testInput = {
            topics: []
        }

        var testExpect = {
            callFailed: false,
            justRecieveNum: 1,
            timeout: 20000
        }

        testSubscribe(testInput, testExpect, done)
    });
});
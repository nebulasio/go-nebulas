'use strict';

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');


var coinbase,
    client,
    server_address;

var env = process.env.NET || 'local';
if (env === 'local') {
    server_address = 'localhost:8684';
    coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
} else if (env === 'testneb1') {
    server_address = '35.182.48.19:8684';
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
} else if (env === "testneb2") {
    server_address = "34.205.26.12:8684";
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
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
'use strict';

var Node = require('../../../node');
var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var nodes = new Node(6);
nodes.Start();

describe('normal transaction', function () {
    before(function (done) {
        this.timeout(6000);
        setTimeout(done, 5000);
    });

    it('normal transfer', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());

        var resp = node.RPC(0).api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce)+1);
        expect(resp).to.be.have.property('txhash');
    });

    it('from & to are same', function () {
        var node = nodes.Node(0);
        // var nebState = node.RPC().api.getNebState();
        var state = node.RPC().api.getAccountState(node.Coinbase());
        // var gasPrice = node.RPC().api.gasPrice();
        // var gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), state.balance, parseInt(state.nonce)+1);
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());

        // sendTransaction
        var resp = node.RPC(0).api.sendTransaction(node.Coinbase(), node.Coinbase(), state.balance, parseInt(state.nonce)+1);
        expect(resp).to.be.have.property('txhash');
        // var nextNebState = node.RPC().api.getNebState();

        // block has been mined
        // while (nextNebState.tail == nebState.tail) {
        //     setTimeout(function(){
        //         console.log('waiting over.');
        //     }, 3000);
        //     nextNebState = node.RPC().api.getNebState();
        // }
        // var nextState = node.RPC().api.getAccountState(node.Coinbase());
        // var oldBalance = new BigNumber(nextState.balance).add(new BigNumber(gas.estimate_gas).mul(new BigNumber(gasPrice.gas_price)));
        // console.log("balance:"+nextState.balance);
        // console.log("new balance:"+oldBalance.toString());
        // expect(nextState).to.be.have.property('balance').eq(oldBalance.toString());
    });

    it('from balance is insufficient', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());

        var value = new BigNumber(state.balance).add("1");
        var resp = node.RPC(0).api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), value.toString(), parseInt(state.nonce)+1);
        expect(resp).to.be.have.property('txhash');
    });

    it('gas is insufficient', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());

        var price = node.RPC().api.gasPrice();
        var gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), state.balance, parseInt(state.nonce)+1);
        gas = new BigNumber(gas.estimate_gas).sub(100);
        var resp = node.RPC(0).api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce)+1, price.gas_price, gas.toString());
        expect(resp).to.be.have.property('txhash');
    });

    it('from is invalid address', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        var resp = node.RPC(0).api.sendTransaction("0x00", nodes.Coinbase(1), state.balance, parseInt(state.nonce)+1);
        // console.log("resp:"+JSON.stringify(resp));
        expect(resp).to.be.have.property('error').equal("address: invalid address");
    });

    it('to is invalid address', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        var resp = node.RPC(0).api.sendTransaction(node.Coinbase(), "0x00", state.balance, parseInt(state.nonce)+1);
        expect(resp).to.be.have.property('error').equal("address: invalid address");
    });

    it('nonce is below', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());

        var resp = node.RPC(0).api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce));
        // console.log("resp:"+JSON.stringify(resp));
        expect(resp).to.be.have.property('error').equal("nonce is invalid");
    });

    it('nonce is heigher', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());

        var nonce = new BigNumber(state.nonce).add(2);
        var resp = node.RPC(0).api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, nonce.toNumber());
        expect(resp).to.be.have.property('txhash');
    });

    it('gasPrice is below', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());

        var price = node.RPC().api.gasPrice();
        var gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), state.balance, parseInt(state.nonce)+1);

        var gasPrice = new BigNumber(price.gas_price).sub(100);
        var resp = node.RPC(0).api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce)+1, gasPrice.toString(), gas.estimate_gas);
        // console.log("resp:"+JSON.stringify(resp));
        expect(resp).to.be.have.property('error').equal("below the gas price");
    });

    it('gas is higher than max', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());

        var price = node.RPC().api.gasPrice();
        var maxGas = new BigNumber(10).pow(9).mul(60);
        var resp = node.RPC(0).api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce)+1, price.gas_price, maxGas.toString());
        // console.log("resp:"+JSON.stringify(resp));
        expect(resp).to.be.have.property('error').equal("out of gas limit");
    });
});

describe('quit', function () {
    it('quit', function () {
        nodes.Stop();
    });
});
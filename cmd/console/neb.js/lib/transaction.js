"use strict";

var protobuf = require('protobufjs');
var utils = require('./utils/utils.js');
var cryptoUtils = require('./utils/crypto-utils.js');
var account = require("./account.js");

var BigNumber = require('bignumber.js');

var SECP256K1 = 1;
var root = protobuf.Root.fromJSON(require("./transaction.json"));

var TxPayloadBinaryType    = "binary";
var TxPayloadDeployType    = "deploy";
var TxPayloadCallType      = "call";
var TxPayloadDelegateType  = "delegate";
var TxPayloadCandidateType = "candidate";

var Transaction = function (chainID, from, to, value, nonce, gasPrice, gasLimit, contract, candidate, delegate) {
    this.chainID = chainID;
    this.from = account.fromAddress(from);
    this.to = account.fromAddress(to);
    this.value = utils.toBigNumber(value);
    this.nonce = nonce;
    this.timestamp =  1516256439;//Math.floor(new Date().getTime()/1000);
    this.data = parsePayload(contract, candidate, delegate);
    this.gasPrice = utils.toBigNumber(gasPrice);
    this.gasLimit = utils.toBigNumber(gasLimit);

    if (this.gasPrice.lessThanOrEqualTo(0)) {
        this.gasPrice = new BigNumber(1000000);
    }

    if (this.gasLimit.lessThanOrEqualTo(0)) {
        this.gasLimit = new BigNumber(20000);
    }
};

var parsePayload = function (contract, candidate, delegate) {
    var payloadType, payload;
    if (utils.isObject(contract) && contract.source.length > 0) {
        payloadType = TxPayloadDeployType;
        payload = {
            SourceType: contract.sourceType,
            Source: contract.source,
            Args: contract.args
        };
    } else if (utils.isObject(contract) && contract.function.length > 0) {
        payloadType = TxPayloadCallType;
        payload = {
            Function: contract.function,
            Args: contract.args
        };
    } else if (utils.isObject(candidate)) {
        payloadType = TxPayloadCandidateType;
        payload = {
            Action: candidate.action
        };
    } else if (utils.isObject(delegate)) {
        payloadType = TxPayloadDelegateType;
        payload = {
            Action: delegate.action,
            Delegatee: delegate.delegatee
        };
    } else {
        payloadType = TxPayloadBinaryType;
        payload = {
            Data: null
        };
    }
    payload = cryptoUtils.toBuffer(JSON.stringify(payload));
    return {type: payloadType, payload: payload};
};

Transaction.prototype = {
    hashTransaction: function () {
        var Data = root.lookup("corepb.Data");
        var err = Data.verify(this.data);
        if (err) {
            throw new Error(err);
        }
        var data = Data.create(this.data);
        var dataBuffer = Data.encode(data).finish();
        var hash = cryptoUtils.sha3(
            this.from.getAddress(),
            this.to.getAddress(),
            cryptoUtils.padToBigEndian(this.value, 128),
            cryptoUtils.padToBigEndian(this.nonce, 64),
            cryptoUtils.padToBigEndian(this.timestamp, 64),
            dataBuffer,
            cryptoUtils.padToBigEndian(this.chainID, 32),
            cryptoUtils.padToBigEndian(this.gasPrice, 128),
            cryptoUtils.padToBigEndian(this.gasLimit, 128)
            );
        return hash;
    },

    signTransaction: function () {
        if (this.from.getPrivateKey() !== null) {
            this.hash = this.hashTransaction();
            this.alg = SECP256K1;
            this.sign = cryptoUtils.sign(this.hash, this.from.getPrivateKey());
        } else {
            throw new Error("transaction from address's private key is invalid");
        }
    },

    toProto: function () {
        var Data = root.lookup("corepb.Data");
        var err = Data.verify(this.data);
        if (err) {
            throw err;
        }
        var data = Data.create(this.data);

        var TransactionProto = root.lookup("corepb.Transaction");

        var txData = {
            hash: this.hash,
            from: this.from.getAddress(),
            to: this.to.getAddress(),
            value: cryptoUtils.toBuffer(this.value),
            nonce: this.nonce,
            timestamp: this.timestamp,
            data: data,
            chainId: this.chainID,
            gasPrice: cryptoUtils.toBuffer(this.gasPrice),
            gasLimit: cryptoUtils.toBuffer(this.gasLimit),
            alg: this.alg,
            sign: this.sign
        };

        err = TransactionProto.verify(txData);
        if (err) {
            throw err;
        }
        var tx = TransactionProto.create(txData);

        var txBuffer = TransactionProto.encode(tx).finish();
        return utils.isBrowser() ? protobuf.util.base64.encode(txBuffer, 0, txBuffer.length) : txBuffer;
    },

    fromProto: function (data) {

        var txBuffer = data;
        if (utils.isBrowser()) {
            txBuffer = new Array(protobuf.util.base64.length(data));
            protobuf.util.base64.decode(data, txBuffer, 0);
        }

        var TransactionProto = root.lookup("corepb.Transaction");
        var txProto = TransactionProto.decode(txBuffer);

        this.hash = cryptoUtils.toBuffer(txProto.hash);
        this.chainID = txProto.chainId;
        this.from = account.fromAddress(txProto.from);
        this.to = account.fromAddress(txProto.to);
        this.value = utils.toBigNumber("0x" + cryptoUtils.toBuffer(txProto.value).toString("hex"));
        // long number is object, should convert to int
        this.nonce = parseInt(txProto.nonce.toString());
        this.timestamp = parseInt(txProto.timestamp.toString());
        this.data = txProto.data;
        this.gasPrice = utils.toBigNumber("0x" + cryptoUtils.toBuffer(txProto.gasPrice).toString("hex"));
        this.gasLimit = utils.toBigNumber("0x" + cryptoUtils.toBuffer(txProto.gasLimit).toString("hex"));

        return this;
    }
};

module.exports = Transaction;

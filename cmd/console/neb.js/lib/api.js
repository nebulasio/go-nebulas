
"use strict";

var utils = require('./utils/utils.js');

var API = function (neb) {
    this._request = neb._request;
};

API.prototype.setRequest = function (request) {
    this._request = request;
};

API.prototype.getNebState = function (callback) {
    return this.request("get", "/v1/user/nebstate", null, callback);
};

API.prototype.nodeInfo = function (callback) {
    return this.request("get", "/v1/user/nodeinfo", null, callback);
};

API.prototype.accounts = function (callback) {
    return this.request("get", "/v1/user/accounts", null, callback);
};

API.prototype.blockDump = function (count, callback) {
    var params = { "count": count };
    return this.request("post", "/v1/user/blockdump", params, callback);
};

API.prototype.getAccountState = function (address, block, callback) {
    var params = { "address": address, "block": block };
    return this.request("post", "/v1/user/accountstate", params, callback);
};

API.prototype.sendTransaction = function (from, to, value, nonce, gasPrice, gasLimit, contract, candidate, delegate, callback) {
    var params = {
        "from": from,
        "to": to,
        "value": utils.toString(value),
        "nonce": nonce,
        "gasPrice": utils.toString(gasPrice),
        "gasLimit": utils.toString(gasLimit),
        "contract": contract,
        "candidate": candidate,
        "delegate": delegate
    };
    return this.request("post", "/v1/user/transaction", params, callback);
};

API.prototype.call = function (from, to, value, nonce, gasPrice, gasLimit, contract, callback) {
    var params = {
        "from": from,
        "to": to,
        "value": utils.toString(value),
        "nonce": nonce,
        "gasPrice": utils.toString(gasPrice),
        "gasLimit": utils.toString(gasLimit),
        "contract": contract
    };
    return this.request("post", "/v1/user/call", params, callback);
};

API.prototype.sendRawTransaction = function (data, callback) {
    var params = { "data": data };
    return this.request("post", "/v1/user/rawtransaction", params, callback);
};

API.prototype.getBlockByHash = function (hash, callback) {
    var params = { "hash": hash };
    return this.request("post", "/v1/user/getBlockByHash", params, callback);
};

API.prototype.getTransactionReceipt = function (hash, callback) {
    var params = { "hash": hash };
    return this.request("post", "/v1/user/getTransactionReceipt", params, callback);
};

API.prototype.subscribe = function (topic, callback) {
    var params = { "topic": topic };
    return this.request("post", "/v1/user/subscribe", params, callback);
};

API.prototype.gasPrice = function (callback) {
    return this.request("get", "/v1/user/getGasPrice", null, callback);
};

API.prototype.estimateGas = function (from, to, value, nonce, gasPrice, gasLimit, contract, candidate, delegate, callback) {
    var params = {
        "from": from,
        "to": to,
        "value": utils.toString(value),
        "nonce": nonce,
        "gasPrice": utils.toString(gasPrice),
        "gasLimit": utils.toString(gasLimit),
        "contract": contract,
        "candidate": candidate,
        "delegate": delegate
    };
    return this.request("post", "/v1/user/estimateGas", params, callback);
};

API.prototype.getEventsByHash = function (hash, callback) {
    var params = { "hash": hash };
    return this.request("post", "/v1/user/getEventsByHash", params, callback);
};

API.prototype.request = function (method, api, params, callback) {
    if (utils.isFunction(callback)) {
        this._request.asyncRequest(method, api, params, callback);
        return callback;
    } else {
        return this._request.request(method, api, params);
    }
};

module.exports = API;


"use strict";

var utils = require('./utils/utils.js');

var API = function (neb) {
    this._request = neb._request;
};

API.prototype.setRequest = function (request) {
    this._request = request;
};

API.prototype.getNebState = function () {
    return this.request("get", "/v1/user/nebstate");
};

API.prototype.nodeInfo = function () {
    return this.request("get", "/v1/user/nodeinfo");
};

API.prototype.accounts = function () {
    return this.request("get", "/v1/user/accounts");
};

API.prototype.getDynasty = function () {
    return this.request("get", "/v1/admin/dynasty");
};

API.prototype.blockDump = function (count) {
    var params = { "count": count };
    return this.request("post", "/v1/user/blockdump", params);
};

API.prototype.getAccountState = function (address, block) {
    var params = { "address": address, "block": block };
    return this.request("post", "/v1/user/accountstate", params);
};

API.prototype.sendTransaction = function (from, to, value, nonce, gasPrice, gasLimit, contract, candidate, delegate) {
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
    return this.request("post", "/v1/user/transaction", params);
};

API.prototype.call = function (from, to, value, nonce, gasPrice, gasLimit, contract) {
    var params = {
        "from": from,
        "to": to,
        "value": utils.toString(value),
        "nonce": nonce,
        "gasPrice": utils.toString(gasPrice),
        "gasLimit": utils.toString(gasLimit),
        "contract": contract
    };
    return this.request("post", "/v1/user/call", params);
};

API.prototype.sendRawTransaction = function (data) {
    var params = { "data": data };
    return this.request("post", "/v1/user/rawtransaction", params);
};

API.prototype.getBlockByHash = function (hash) {
    var params = { "hash": hash };
    return this.request("post", "/v1/user/getBlockByHash", params);
};

API.prototype.getTransactionReceipt = function (hash) {
    var params = { "hash": hash };
    return this.request("post", "/v1/user/getTransactionReceipt", params);
};

API.prototype.subscribe = function (topic) {
    var params = { "topic": topic };
    return this.request("post", "/v1/user/subscribe", params);
};

API.prototype.gasPrice = function () {
    return this.request("get", "/v1/user/getGasPrice");
};

API.prototype.estimateGas = function (from, to, value, nonce, gasPrice, gasLimit, contract, candidate, delegate) {
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
    return this.request("post", "/v1/user/estimateGas", params);
};

API.prototype.getEventsByHash = function (hash) {
    var params = { "hash": hash };
    return this.request("post", "/v1/user/getEventsByHash", params);
};

API.prototype.request = function (method, api, params) {
    return this._request.request(method, api, params);
};

module.exports = API;

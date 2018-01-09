
"use strict";

var utils = require('./utils/utils.js');

var Admin = function (neb) {
    this._request = neb._request;
};

Admin.prototype.setRequest = function (request) {
    this._request = request;
};

Admin.prototype.newAccount = function (passphrase, callback) {
    var params = { "passphrase": passphrase };
    return this.request("post", "/v1/admin/account/new", params, callback);
};

Admin.prototype.unlockAccount = function (address, passphrase, callback) {
    var params = {
        "address": address,
        "passphrase": passphrase
    };
    return this.request("post", "/v1/admin/account/unlock", params, callback);
};

Admin.prototype.lockAccount = function (address, callback) {
    var params = { "address": address };
    return this.request("post", "/v1/admin/account/lock", params, callback);
};

Admin.prototype.changeNetworkID = function (networkId, callback) {
    var params = { "networkId": networkId };
    return this.request("post", "/v1/admin/changeNetworkID", params, callback);
};

Admin.prototype.signTransaction = function (from, to, value, nonce, gasPrice, gasLimit, contract, candidate, delegate, callback) {
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
    return this.request("post", "/v1/admin/sign", params, callback);
};

Admin.prototype.sendTransactionWithPassphrase = function (from, to, value, nonce, gasPrice, gasLimit, contract, candidate, delegate, passphrase, callback) {
    var tx = {
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
    var params = {
        "transaction": tx,
        "passphrase": passphrase
    };
    return this.request("post", "/v1/admin/transactionWithPassphrase", params, callback);
};

Admin.prototype.getDynasty = function (callback) {
    return this.request("get", "/v1/admin/dynasty", null, callback);
};

Admin.prototype.getDelegateVoters = function (delegatee, callback) {
    var params = { "delegatee": delegatee };
    return this.request("post", "/v1/admin/delegateVoters", params, callback);
};

Admin.prototype.startMine = function (passphrase, callback) {
    var params = { "passphrase": passphrase };
    return this.request("post", "/v1/admin/startMine", params, callback);
};

Admin.prototype.stopMine = function (callback) {
    return this.request("get", "/v1/admin/stopMine", null, callback);
};

Admin.prototype.request = function (method, api, params, callback) {
    if (utils.isFunction(callback)) {
        this._request.asyncRequest(method, api, params, callback);
        return callback;
    } else {
        return this._request.request(method, api, params);
    }
};

module.exports = Admin;


"use strict";

var utils = require('./utils/utils.js');

var Admin = function (neb) {
	this._request = neb._request;
};

Admin.prototype.setRequest = function (request) {
	this._request = request;
};

Admin.prototype.newAccount = function (passphrase) {
	var params = {"passphrase": passphrase};
	return this.request("post", "/v1/admin/account/new", params);
};

Admin.prototype.unlockAccount = function (address, passphrase) {
	var params = {"address": address,
	 "passphrase": passphrase};
	return this.request("post", "/v1/admin/account/unlock", params);
};

Admin.prototype.lockAccount = function (address) {
	var params = {"address": address};
	return this.request("post", "/v1/admin/account/lock", params);
};

Admin.prototype.signTransaction = function (from, to, value, nonce, gasPrice, gasLimit, contract) {
    var params = {"from": from,
        "to": to,
        "value": utils.toString(value),
        "nonce": nonce,
        "gasPrice": utils.toString(gasPrice),
        "gasLimit": utils.toString(gasLimit),
        "contract": contract
    };
	return this.request("post", "/v1/admin/sign", params);
};

Admin.prototype.sendTransactionWithPassphrase = function (from, to, value, nonce, gasPrice, gasLimit, contract, passphrase) {
    var tx = {"from": from,
        "to": to,
        "value": utils.toString(value),
        "nonce": nonce,
        "gasPrice": utils.toString(gasPrice),
        "gasLimit": utils.toString(gasLimit),
        "contract": contract
    };
    var params = {"transaction": tx,
        "passphrase": passphrase
	};
	return this.request("post", "/v1/admin/transactionWithPassphrase", params);
};

Admin.prototype.request = function (method, api, params) {
	return this._request.request(method, api, params);
};

module.exports = Admin;

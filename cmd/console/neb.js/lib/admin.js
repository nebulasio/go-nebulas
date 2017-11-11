
"use strict";

var Admin = function (neb) {
	this._requestHandler = neb._requestHandler;
};

Admin.prototype.newAccount = function (passphrase) {
	var params = {"passphrase": passphrase};
	return this.request("get", "/v1/newAccount", params);
};

Admin.prototype.unlockAccount = function (address, passphrase) {
	var params = {"address": address,
	 "passphrase": passphrase};
	return this.request("post", "/v1/unlock", params);
};

Admin.prototype.lockAccount = function (address) {
	var params = {"address": address};
	return this.request("post", "/v1/lock", params);
};

Admin.prototype.signTransaction = function (from, to, value, nonce, source, args) {
	var params = {"from": from,
	"to": to,
	"value": value,
	"nonce": nonce,
	"source": source,
	"args": args
	};
	return this.request("post", "/v1/sign", params);
};

Admin.prototype.sendTransactionWithPassphrase = function (from, to, value, nonce, source, args, passphrase) {
	var params = {"from": from,
	"to": to,
	"value": value,
	"nonce": nonce,
	"source": source,
	"args": args,
	"passphrase": passphrase
	};
	return this.request("post", "/v1/transactionWithPassphrase", params);
};

Admin.prototype.request = function (method, api, params) {
	return this._requestHandler.request(method, api, params);
};

module.exports = Admin;
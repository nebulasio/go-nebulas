
"use strict";

var API = function (neb) {
	this._requestHandler = neb._requestHandler;
};

API.prototype.getNebState = function () {
	return this.request("get", "/v1/neb/state");
};

API.prototype.nodeInfo = function () {
	return this.request("get", "/v1/node/info");
};

API.prototype.accounts = function () {
	return this.request("get", "/v1/accounts");
};

API.prototype.blockDump = function (count) {
	var params = {"count":count};
	return this.request("post", "/v1/block/dump", params);
};

API.prototype.getAccountState = function (address) {
	var params = {"address":address};
	return this.request("post", "/v1/account/state", params);
};

API.prototype.sendTransaction = function (from, to, value, nonce, source, args, gasPrice, gasLimit) {
	var params = {"from": from,
	"to": to,
	"value": value,
	"nonce": nonce,
	"source": source,
	"args": args,
	"gasPrice": gasPrice,
	"gasLimit": gasLimit
	};
	return this.request("post", "/v1/transaction", params);
};

API.prototype.call = function (from, to, nonce, func, args, gasPrice, gasLimit) {
	var params = {"from": from,
	"to": to,
	"nonce": nonce,
	"function": func,
	"args": args,
	"gasPrice": gasPrice,
	"gasLimit": gasLimit
	};
	return this.request("post", "/v1/call", params);
};

API.prototype.sendRawTransaction = function (data) {
	var params = {"data": data};
	return this.request("post", "/v1/rawtransaction", params);
};

API.prototype.getBlockByHash = function (hash) {
	var params = {"hash": hash};
	return this.request("post", "/v1/getBlockByHash", params);
};

API.prototype.getTransactionReceipt = function (hash) {
	var params = {"hash": hash};
	return this.request("post", "/v1/getTransactionReceipt", params);
};

API.prototype.request = function (method, api, params) {
	return this._requestHandler.request(method, api, params);
};

module.exports = API;
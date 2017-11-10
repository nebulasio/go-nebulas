
"use strict";

var API = function (neb) {
	this._requestHandler = neb.requestHandler;
};

API.prototype.getNebState = function () {
	return this._request("get", "/v1/neb/state");
};

API.prototype.nodeInfo = function () {
	return this._request("get", "/v1/node/info");
};

API.prototype.accounts = function () {
	return this._request("get", "/v1/accounts");
};

API.prototype.blockDump = function (count) {
	var params = {"count":count};
	return this._request("post", "/v1/block/dump", params);
};

API.prototype.getAccountState = function (address) {
	var params = {"address":address};
	return this._request("post", "/v1/account/state", params);
};

API.prototype.sendTransaction = function (from, to, value, nonce, source, args) {
	var params = {"from": from,
	"to": to,
	"value": value,
	"nonce": nonce,
	"source": source,
	"args": args
	};
	return this._request("post", "/v1/transaction", params);
};

API.prototype.call = function (from, to, nonce, func, args) {
	var params = {"from": from,
	"to": to,
	"nonce": nonce,
	"function": func,
	"args": args
	};
	return this._request("post", "/v1/call", params);
};

API.prototype.sendRawTransaction = function (data) {
	var params = {"data": data};
	return this._request("post", "/v1/rawtransaction", params);
};

API.prototype.getBlockByHash = function (hash) {
	var params = {"hash": hash};
	return this._request("post", "/v1/getBlockByHash", params);
};

API.prototype.getTransactionReceipt = function (hash) {
	var params = {"hash": hash};
	return this._request("post", "/v1/getTransactionReceipt", params);
};

API.prototype._request = function (method, api, params) {
	return this._requestHandler.request(method, api, params);
};

module.exports = API;
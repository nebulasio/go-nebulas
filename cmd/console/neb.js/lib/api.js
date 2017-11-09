
"use strict";

var API = function (neb) {
	this._requestHandler = neb.requestHandler;
};

API.prototype.getNebState = function () {
	return this._request("get", "/v1/neb/state");
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

API.prototype.sendTransaction = function (from, to, value, nonce, source, func, args) {
	var params = {"from": from,
	"to": to,
	"value": value,
	"nonce": nonce,
	"source": source,
	"function": func,
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

API.prototype._request = function (method, api, params) {
	return this._requestHandler.request(method, api, params);
};

module.exports = API;
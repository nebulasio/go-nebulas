// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

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
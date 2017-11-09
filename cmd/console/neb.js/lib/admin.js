
"use strict";

var Admin = function (neb) {
	this._requestHandler = neb.requestHandler;
};

Admin.prototype._request = function (method, api, params) {
	return this._requestHandler.request(method, api, params);
};

module.exports = Admin;
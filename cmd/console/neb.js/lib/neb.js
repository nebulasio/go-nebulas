
"use strict";

var HttpRequest = require("./httprequest.js");

var API = require("./api.js");
var Admin = require("./admin.js");

var Neb = function (request) {
	if (request) {
		this._requestHandler = request;
	} else {
		this._requestHandler = new HttpRequest();
	}

	this.api = new API(this);
	this.admin = new Admin(this);
};

Neb.prototype.setRequestHandler = function (request) {
	this._requestHandler = request;
};

module.exports = Neb;
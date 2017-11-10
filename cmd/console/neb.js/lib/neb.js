
"use strict";

var HttpRequest = require("./httprequest.js");

var API = require("./api.js");
var Admin = require("./admin.js");

var Neb = function (request) {
	if (request) {
		this.requestHandler = request;
	} else {
		this.requestHandler = new HttpRequest();
	}

	this.api = new API(this);
	this.admin = new Admin(this);
};

Neb.prototype.setRequestHandler = function (request) {
	this.requestHandler = request;
};

module.exports = Neb;
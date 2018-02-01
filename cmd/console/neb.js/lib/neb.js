
"use strict";

var HttpRequest = require("./httprequest.js");

var API = require("./api.js");
var Admin = require("./admin.js");

var Unit = require("./utils/unit.js");

var Neb = function (request) {
	if (request) {
		this._request = request;
	} else {
		this._request = new HttpRequest();
	}

	this.api = new API(this);
	this.admin = new Admin(this);
};

Neb.prototype.setRequest = function (request) {
	this._request = request;
	this.api.setRequest(request);
	this.admin.setRequest(request);
};

Neb.HttpRequest = HttpRequest;

Neb.prototype.toBasic = Unit.toBasic;
Neb.prototype.fromBasic = Unit.fromBasic;
Neb.prototype.nasToBasic = Unit.nasToBasic;

module.exports = Neb;

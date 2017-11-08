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

var XMLHttpRequest = null;

// browser
if (typeof window !== 'undefined' && window.XMLHttpRequest) {
  XMLHttpRequest = window.XMLHttpRequest;
// node
} else {
  XMLHttpRequest = require("xmlhttprequest").XMLHttpRequest;
}

var XHR2 = require('xhr2'); 

var HttpRequest = function (host, port,timeout) {
	this.host = host || "http://localhost";
	this.port = port || "8080";
	this.timeout = timeout || 0;
};

HttpRequest.prototype._newRequest = function (method, api, async) {
	var request;
	if (async) {
		request = new XHR2();
		request.timeout = this.timeout;
	} else {
		request = new XMLHttpRequest();
	}

	var m = method.toUpperCase() === "POST" ? "POST" : "GET";
	var url = this.host + ":" + this.port + api;
	request.open(m, url, async);
	return request;
};

HttpRequest.prototype.request = function (method, api, payload) {
	var request = this._newRequest(method, api, false);
	try {
		if (payload === undefined || payload === "") {
			request.send();
		} else {
			request.send(JSON.stringify(payload));
		}
	} catch (error) {
		throw error;
	}

	var result = request.responseText;
	try {
		result = JSON.parse(result);
	} catch (e) {
		throw e;
	}

	return result;
};

HttpRequest.prototype.asyncRequest = function (method, api, payload, callback) {
	var request = this._newRequest(method, api, true);
	request.onreadystatechange = function () {
	    if (request.readyState === 4 && request.timeout !== 1) {
	      var result = request.responseText;
	      var error = null;

	      try {
	        result = JSON.parse(result);
	      } catch (e) {
	        var message = !!result && !!result.error && !!result.error.message ? result.error.message : 'Invalid response: ' + JSON.stringify(result);
        	error = new Error(message);
	      }

	      callback(error, result);
	    }
  };

  request.ontimeout = function () {
    callback(new Error("connection timeout"));
  };

  try {
  	if (payload === undefined || payload === "") {
		request.send();
	} else {
		request.send(JSON.stringify(payload));
	}
  } catch (error) {
    callback(error);
  }
};

module.exports = HttpRequest;

"use strict";

// browser
if (typeof window !== "undefined" && window.XMLHttpRequest) {
  XMLHttpRequest = window.XMLHttpRequest; // jshint ignore: line
// node
} else {
  XMLHttpRequest = require("xmlhttprequest").XMLHttpRequest; // jshint ignore: line
}

var XHR2 = require("xhr2"); 

var HttpRequest = function (host, timeout) {
	this.host = host || "http://localhost:8191";
	this.timeout = timeout || 0;
};

HttpRequest.prototype.setHost = function (host) {
	this.host = host || "http://localhost:8191";
};

HttpRequest.prototype._newRequest = function (method, api, async) {
	var request;
	if (async) {
		request = new XHR2();
		request.timeout = this.timeout;
	} else {
		request = new XMLHttpRequest();
	}
	var m = "GET";
	if (method.toUpperCase() === "POST") {
		m = "POST";
	}
	var url = this.host + api;
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
	        var message = !!result && !!result.error && !!result.error.message ? result.error.message : "Invalid response: " + JSON.stringify(result);
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
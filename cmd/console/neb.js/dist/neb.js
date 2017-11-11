require=(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){

"use strict";

var Admin = function (neb) {
	this._requestHandler = neb._requestHandler;
};

Admin.prototype.newAccount = function (passphrase) {
	var params = {"passphrase": passphrase};
	return this.request("get", "/v1/newAccount", params);
};

Admin.prototype.unlockAccount = function (address, passphrase) {
	var params = {"address": address,
	 "passphrase": passphrase};
	return this.request("post", "/v1/unlock", params);
};

Admin.prototype.lockAccount = function (address) {
	var params = {"address": address};
	return this.request("post", "/v1/lock", params);
};

Admin.prototype.signTransaction = function (from, to, value, nonce, source, args) {
	var params = {"from": from,
	"to": to,
	"value": value,
	"nonce": nonce,
	"source": source,
	"args": args
	};
	return this.request("post", "/v1/sign", params);
};

Admin.prototype.sendTransactionWithPassphrase = function (from, to, value, nonce, source, args, passphrase) {
	var params = {"from": from,
	"to": to,
	"value": value,
	"nonce": nonce,
	"source": source,
	"args": args,
	"passphrase": passphrase
	};
	return this.request("post", "/v1/transactionWithPassphrase", params);
};

Admin.prototype.request = function (method, api, params) {
	return this._requestHandler.request(method, api, params);
};

module.exports = Admin;
},{}],2:[function(require,module,exports){

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

API.prototype.sendTransaction = function (from, to, value, nonce, source, args) {
	var params = {"from": from,
	"to": to,
	"value": value,
	"nonce": nonce,
	"source": source,
	"args": args
	};
	return this.request("post", "/v1/transaction", params);
};

API.prototype.call = function (from, to, nonce, func, args) {
	var params = {"from": from,
	"to": to,
	"nonce": nonce,
	"function": func,
	"args": args
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
},{}],3:[function(require,module,exports){

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
	this.host = host || "http://localhost:8080";
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
},{"xhr2":6,"xmlhttprequest":5}],4:[function(require,module,exports){

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
},{"./admin.js":1,"./api.js":2,"./httprequest.js":3}],5:[function(require,module,exports){
"use strict";

// go env doesn"t have and need XMLHttpRequest
if (typeof XMLHttpRequest === "undefined") {
    exports.XMLHttpRequest = {};
} else {
    exports.XMLHttpRequest = XMLHttpRequest; // jshint ignore:line
}


},{}],6:[function(require,module,exports){
module.exports = XMLHttpRequest;

},{}],"neb":[function(require,module,exports){
var Neb = require('./lib/neb');

// dont override global variable
if (typeof window !== 'undefined' && typeof window.Neb === 'undefined') {
    window.Neb = Neb;
}

module.exports = Neb;

},{"./lib/neb":4}]},{},["neb"])
//# sourceMappingURL=neb.js.map

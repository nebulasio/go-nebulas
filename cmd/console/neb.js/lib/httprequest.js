"use strict";

// must run: npm install promise
// Node
var Promise = require('promise/lib/es6-extensions');
var XMLHttpRequest;

// browser
if (typeof window !== "undefined" && window.XMLHttpRequest) {
    XMLHttpRequest = window.XMLHttpRequest;

    // node
} else {
    XMLHttpRequest = require("xmlhttprequest").XMLHttpRequest;

}

var HttpRequest = function (host, timeout) {
    this.host = host || "http://localhost:8685";
    this.timeout = timeout || 0;
};

HttpRequest.prototype.setHost = function (host) {
    this.host = host || "http://localhost:8685";
};

HttpRequest.prototype._newRequest = function (method, api, async) {
    var request = new XMLHttpRequest();
    var m = "GET";
    if (method.toUpperCase() === "POST") {
        m = "POST";
    }
    var url = this.host + api;
    request.open(m, url, async);
    return request;
};

HttpRequest.prototype.request = function (method, api, payload) {
    var request = this._newRequest(method, api, true);

    var promise = new Promise(function (resolve, reject) {

        request.onload = function () {
            if (request.readyState === 4 && request.timeout !== 1) {

                resolve(request.responseText);

            }else{
                reject(Error(request.statusText));
            }
        };

        request.onerror = function () {
            reject(Error("Network Error"));
        };

        request.ontimeout = function () {
            reject(new Error("connection timeout"));
        };

        try {
            if (payload === undefined || payload === "") {
                request.send();
            } else {
                request.send(JSON.stringify(payload));
            }
        } catch (error) {
            reject(error);
        }
    });

    return promise;
};

HttpRequest.prototype.asyncRequest = function (method, api, payload, callback) {
    this.request(method, api, payload).then(function (resp) {
        callback(null, resp);
    }).catch(function (err) {
        callback(err);
    });
};

module.exports = HttpRequest;

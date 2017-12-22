
"use strict";

var Request = require("sync-request");

var HttpRequest = function (host, timeout) {
    this.host = host || "http://localhost:8090";
    this.timeout = timeout || 0;
};

HttpRequest.prototype.setHost = function (host) {
    this.host = host || "http://localhost:8090";
};

HttpRequest.prototype._newRequest = function (method, api, payload) {
    var m = "GET";
    if (method.toUpperCase() === "POST") {
        m = "POST";
    }
    var url = this.host + api;
    var resp = Request(m, url, { json: payload});
    return resp.body;
};

HttpRequest.prototype.request = function (method, api, payload) {
    var result = this._newRequest(method, api, payload);
    try {
        result = JSON.parse(result);
    } catch (e) {
        throw e;
    }

    return result;
};

HttpRequest.prototype.asyncRequest = function (method, api, payload, callback) {
    var result = this._newRequest(method, api, payload);
    try {
        result = JSON.parse(result);
    } catch (e) {
        callback(error, result);
    }
    callback(null, result);
};

module.exports = HttpRequest;

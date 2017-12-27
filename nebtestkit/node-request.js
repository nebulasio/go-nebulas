
"use strict";

var rq = require("request-promise");

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
    var options = {
        method: m,
        uri: url,
        body: payload,
        json: true // Automatically stringifies the body to JSON
    };
    return options;
};

HttpRequest.prototype.request = function (method, api, payload) {
    var options = this._newRequest(method, api, payload);
    return rq(options);
};

HttpRequest.prototype.asyncRequest = function (method, api, payload, callback) {
    var options = this._newRequest(method, api, payload);
    rq(options).then(function (resp) {
        callback(null, resp);
    }).catch(function (err) {
        callback(err, null);
    });
};

module.exports = HttpRequest;

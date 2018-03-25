
"use strict";

var rq = require("request-promise");

var HttpRequest = function (host, timeout, apiVersion) {
    this.host = host || "http://localhost:8685";
    this.timeout = timeout || 0;
    this.APIVersion = apiVersion || 'v1';
};

HttpRequest.prototype.setHost = function (host) {
    this.host = host || "http://localhost:8685";
};

HttpRequest.prototype.setAPIVersion = function (APIVersion) {
    this.APIVersion = APIVersion;
};

HttpRequest.prototype.createUrl = function (action) {
    return this.host + '/' + this.APIVersion + action;
};

HttpRequest.prototype._newRequest = function (method, api, payload) {
    var m = "GET";
    if (method.toUpperCase() === "POST") {
        m = "POST";
    }
    var url = this.createUrl(api);
    var options = {
        method: m,
        uri: url,
        body: payload,
        json: true, // Automatically stringifies the body to JSON
        transform: function (body, response, resolveWithFullResponse) {
            return body.result;
        }
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

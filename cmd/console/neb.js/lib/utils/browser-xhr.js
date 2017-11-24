"use strict";

if (typeof XMLHttpRequest === "undefined") {
    exports.XMLHttpRequest = {};
} else {
    exports.XMLHttpRequest = XMLHttpRequest; // jshint ignore:line
}


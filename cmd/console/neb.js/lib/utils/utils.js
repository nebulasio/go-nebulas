
"use strict";

var BigNumber = require('bignumber.js');

var isNull = function (v) {
    return v === null || typeof v === "undefined";
};

var isBrowser = function () {
    return (typeof window !== "undefined");
};

var isBigNumber = function (obj) {
    return obj instanceof BigNumber ||
        (obj && obj.constructor && obj.constructor.name === 'BigNumber');
};

var isString = function (obj) {
    return typeof obj === 'string' && obj.constructor === String;
};

var isObject = function (obj) {
    return obj !== null && typeof obj === 'object';
};

var isFunction = function (object) {
    return typeof object === 'function';
};

var toBigNumber = function (number) {
	number = number || 0;
	if (isBigNumber(number)) {
		return number;
	}
	if (isString(number) && number.indexOf('0x') === 0) {
        return new BigNumber(number.replace('0x',''), 16);
    }
    return new BigNumber(number.toString(10), 10);
};

var toString = function (obj) {
	if (isString(obj)) {
		return obj;
	} else if (isBigNumber(obj)) {
		return obj.toString(10);
	} else if (isObject(obj)) {
		return JSON.stringify(obj);
	} else {
		return obj + "";
	}
};

module.exports = {
    isNull: isNull,
    isBrowser: isBrowser,
	isBigNumber: isBigNumber,
	isString: isString,
	isObject: isObject,
    isFunction: isFunction,
	toBigNumber: toBigNumber,
	toString: toString
};

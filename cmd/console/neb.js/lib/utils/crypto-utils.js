
"use strict";

var Buffer = require('safe-buffer').Buffer;

var sha3256 = require('js-sha3').sha3_256; // jshint ignore: line
var keccak256 = require('js-sha3').keccak256;
var secp256k1 = require('secp256k1');
var crypto = require('crypto');
var scrypt = require('scryptsy');

var uuid = require('uuid');

var assert = require('assert');

var utils = require('./utils.js');

var sha3 = function (v) {
    v = toBuffer(v);
    v = v.toString("hex");
    return Buffer.from(sha3256(v), "hex");
};

// check if hex string
var isHexPrefixed = function (str) {
    if (typeof str !== 'string') {
        throw new Error("[is-hex-prefixed] value must be type 'string', is currently type " + (typeof str) + ", while checking isHexPrefixed.");
    }

    return str.slice(0, 2) === '0x';
};

// returns hex string without 0x
var stripHexPrefix = function (str) {
    if (typeof str !== 'string') {
        return str;
    }
    return isHexPrefixed(str) ? str.slice(2) : str;
};

function isHexString(value, length) {
    if (typeof(value) !== 'string' || !value.match(/^0x[0-9A-Fa-f]*$/)) {
        return false;
    }

    if (length && value.length !== 2 + 2 * length) { return false; }

    return true;
}

// returns hex string from int
function intToHex(i) {
    var hex = i.toString(16); // eslint-disable-line

    return '0x' + padToEven(hex);
}

// returns buffer from int
function intToBuffer(i) {
    var hex = intToHex(i);

    return new Buffer(hex.slice(2), 'hex');
}

// returns a buffer filled with 0
var zeros = function (bytes) {
    return Buffer.allocUnsafe(bytes).fill(0);
};

var padToEven = function (value) {
    var a = value; // eslint-disable-line

    if (typeof a !== 'string') {
        throw new Error('padToEven only support string');
    }

    if (a.length % 2) {
        a = '0' + a;
    }

    return a;
};

// attempts to turn a value to buffer, the input can be buffer, string,number
var toBuffer = function (v) {
    /*jshint maxcomplexity:9 */

    if (!Buffer.isBuffer(v)) {
        if (Array.isArray(v)) {
            v = Buffer.from(v);
        } else if (typeof v === 'string') {
            if (isHexString(v)) {
                v = Buffer.from(padToEven(stripHexPrefix(v)), 'hex');
            } else {
                v = Buffer.from(v);
            }
        } else if (typeof v === 'number') {
            v = intToBuffer(v);
        } else if (v === null || v === undefined) {
            v = Buffer.allocUnsafe(0);
        } else if (utils.isBigNumber(v)) {
            v = Buffer.from(padToEven(v.toString(16)), 'hex');
        } else {
            throw new Error('invalid type');
        }
    }
    return v;
};

var bufferToHex = function (buf) {
    buf = toBuffer(buf);
    return '0x' + buf.toString('hex');
};

// returns address from private key
var privateToAddress = function (privateKey) {
    return publicToAddress(privateToPublic(privateKey));
};

// convert secp256k1 private key to public key
var privateToPublic = function (privateKey) {
    privateKey = toBuffer(privateKey);
    // skip the type flag and use the X, Y points
    return secp256k1.publicKeyCreate(privateKey, false).slice(1);
};

var publicToAddress = function (pubKey, sanitize) {
    pubKey = toBuffer(pubKey);
    if (sanitize && (pubKey.length !== 64)) {
        pubKey = secp256k1.publicKeyConvert(pubKey, false).slice(1);
    }
    assert(pubKey.length === 64);

    // Only take the lower 160bits of the hash
    var content = sha3(pubKey).slice(-20);
    var checksum = sha3(content).slice(0,4);
    return Buffer.concat([content, checksum]);
};

var isValidPublic = function (publicKey, sanitize) {
    if (publicKey.length === 64) {
        // Convert to SEC1 for secp256k1
        return secp256k1.publicKeyVerify(Buffer.concat([ Buffer.from([4]), publicKey ]));
    }

    if (!sanitize) {
        return false;
    }

    return secp256k1.publicKeyVerify(publicKey);
};



module.exports = {
    secp256k1: secp256k1,
    keccak256: keccak256,
    sha3: sha3,
    crypto: crypto,
    scrypt: scrypt,
    uuid: uuid,

    zeros: zeros,
    toBuffer: toBuffer,
    bufferToHex: bufferToHex,
    privateToAddress: privateToAddress,
    privateToPublic: privateToPublic,
    publicToAddress: publicToAddress,
    isValidPublic: isValidPublic
};

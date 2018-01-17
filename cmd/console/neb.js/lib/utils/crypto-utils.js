
"use strict";

var Buffer = require('safe-buffer').Buffer;

var jsSHA = require('jssha');
var createKeccakHash = require('keccak');
var secp256k1 = require('secp256k1');
var crypto = require('crypto');
var scrypt = require('scryptsy');

var uuid = require('uuid');

var utils = require('./utils.js');

var keccak = function (a, bits) {
    a = toBuffer(a);
    if (!bits) bits = 256;

    return createKeccakHash('keccak' + bits).update(a).digest();
};

var sha3 = function () {
    var shaObj = new jsSHA("SHA3-256", "HEX");
    for (var i = 0; i < arguments.length; i++) {
        var v = toBuffer(arguments[i]);
        shaObj.update(v.toString("hex"));
    }
    return Buffer.from(shaObj.getHash("HEX"), "hex");
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

// convert value to digit/8 buffer with BigEndian.
var padToBigEndian = function (value, digit) {
    value = toBuffer(value);
    var buff = Buffer.alloc(digit/8);
    for (var i = 0; i < value.length; i++) {
        var start = buff.length - value.length + i;
        if ( start >= 0) {
            buff[start] = value[i];
        }
    }
    return buff;
};

// attempts to turn a value to buffer, the input can be buffer, string,number
var toBuffer = function (v) {
    /*jshint maxcomplexity:13 */

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
            // TODO: neb number is a big int, not support if v is decimal, later fix it.
            v = Buffer.from(padToEven(v.toString(16)), 'hex');
        } else if (v.toArray) {
            v = Buffer.from(v.toArray());
        } else if (v.subarray) {
            v = Buffer.from(v);
        } else if (v === null || typeof v === "undefined") {
            v = Buffer.allocUnsafe(0);
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

// convert secp256k1 private key to public key
var privateToPublic = function (privateKey) {
    privateKey = toBuffer(privateKey);
    // skip the type flag and use the X, Y points
    return secp256k1.publicKeyCreate(privateKey, false).slice(1);
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

// sign transaction hash
var sign = function (msgHash, privateKey) {

    var sig = secp256k1.sign(toBuffer(msgHash), toBuffer(privateKey));
    // var ret = {}
    // ret.r = sig.signature.slice(0, 32)
    // ret.s = sig.signature.slice(32, 64)
    // ret.v = sig.recovery
    return Buffer.concat([toBuffer(sig.signature), toBuffer(sig.recovery)]);
};

module.exports = {
    secp256k1: secp256k1,
    keccak: keccak,
    sha3: sha3,
    crypto: crypto,
    scrypt: scrypt,
    uuid: uuid,

    zeros: zeros,
    isHexPrefixed: isHexPrefixed,
    padToBigEndian: padToBigEndian,
    toBuffer: toBuffer,
    bufferToHex: bufferToHex,
    privateToPublic: privateToPublic,
    isValidPublic: isValidPublic,
    sign: sign
};

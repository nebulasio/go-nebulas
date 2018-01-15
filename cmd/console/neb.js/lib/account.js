
"use strict";

var cryptoUtils = require('./utils/crypto-utils.js');

var Account = function (priv, path) {
    if (typeof priv !== "undefined") {
        this.privKey = priv.length === 32 ? priv : Buffer(priv, 'hex');
    }
    this.path = path;
};

Account.NewAccount = function () {
    return new Account(cryptoUtils.crypto.randomBytes(32));
    // return new Account(new Buffer("ac3773e06ae74c0fa566b0e421d4e391333f31aef90b383f0c0e83e4873609d6", "hex"))
};

Account.prototype = {

    getPrivateKey: function () {
        return this.privKey;
    },

    getPrivateKeyString: function () {
        if (typeof this.privKey !== "undefined") {
            return this.getPrivateKey().toString('hex');
        } else {
            return "";
        }
    },

    getPublicKey: function () {
        if (typeof this.pubKey === "undefined") {
            this.pubKey = cryptoUtils.privateToPublic(this.privKey);
        }
        return this.pubKey;
    },

    getPublicKeyString: function () {
        return this.getPublicKey().toString('hex');
    },

    getAddress: function () {
        return cryptoUtils.publicToAddress(this.getPublicKey(), true);
    },

    getAddressString: function () {
        return this.getAddress().toString('hex');
    },

    toKey: function (password, opts) {
        /*jshint maxcomplexity:16 */

        opts = opts || {};
        var salt = opts.salt || cryptoUtils.crypto.randomBytes(32);
        var iv = opts.iv || cryptoUtils.crypto.randomBytes(16);
        var derivedKey;
        var kdf = opts.kdf || 'scrypt';
        var kdfparams = {
            dklen: opts.dklen || 32,
            salt: salt.toString('hex')
        };
        if (kdf === 'pbkdf2') {
            kdfparams.c = opts.c || 262144;
            kdfparams.prf = 'hmac-sha256';
            derivedKey = cryptoUtils.crypto.pbkdf2Sync(new Buffer(password), salt, kdfparams.c, kdfparams.dklen, 'sha256');
        } else if (kdf === 'scrypt') {
            kdfparams.n = opts.n || 262144;
            kdfparams.r = opts.r || 8;
            kdfparams.p = opts.p || 1;
            derivedKey = cryptoUtils.scrypt(new Buffer(password), salt, kdfparams.n, kdfparams.r, kdfparams.p, kdfparams.dklen);
        } else {
            throw new Error('Unsupported kdf');
        }
        var cipher = cryptoUtils.crypto.createCipheriv(opts.cipher || 'aes-128-ctr', derivedKey.slice(0, 16), iv);
        if (!cipher) {
            throw new Error('Unsupported cipher');
        }
        var ciphertext = Buffer.concat([cipher.update(this.privKey), cipher.final()]);
        var mac = cryptoUtils.sha3(Buffer.concat([derivedKey.slice(16, 32), new Buffer(ciphertext, 'hex')]));
        return {
            version: 3,
            id: cryptoUtils.uuid.v4({
                random: opts.uuid || cryptoUtils.crypto.randomBytes(16)
            }),
            address: this.getAddress().toString('hex'),
            crypto: {
                ciphertext: ciphertext.toString('hex'),
                cipherparams: {
                    iv: iv.toString('hex')
                },
                cipher: opts.cipher || 'aes-128-ctr',
                kdf: kdf,
                kdfparams: kdfparams,
                mac: mac.toString('hex')
            }
        };
    },

    toKeyString: function (password, opts) {
        return JSON.stringify(this.toKey(password, opts));
    },

    fromKey: function (input, password, nonStrict) {
        /*jshint maxcomplexity:9 */

        var json = (typeof input === 'object') ? input : JSON.parse(nonStrict ? input.toLowerCase() : input);
        if (json.version !== 3) {
            throw new Error('Not a V3 wallet');
        }
        var derivedKey;
        var kdfparams;
        if (json.crypto.kdf === 'scrypt') {
            kdfparams = json.crypto.kdfparams;
            derivedKey = cryptoUtils.scrypt(new Buffer(password), new Buffer(kdfparams.salt, 'hex'), kdfparams.n, kdfparams.r, kdfparams.p, kdfparams.dklen);
        } else if (json.crypto.kdf === 'pbkdf2') {
            kdfparams = json.crypto.kdfparams;
            if (kdfparams.prf !== 'hmac-sha256') {
                throw new Error('Unsupported parameters to PBKDF2');
            }
            derivedKey = cryptoUtils.crypto.pbkdf2Sync(new Buffer(password), new Buffer(kdfparams.salt, 'hex'), kdfparams.c, kdfparams.dklen, 'sha256');
        } else {
            throw new Error('Unsupported key derivation scheme');
        }
        var ciphertext = new Buffer(json.crypto.ciphertext, 'hex');
        var mac = cryptoUtils.sha3(Buffer.concat([derivedKey.slice(16, 32), ciphertext]));
        if (mac.toString('hex') !== json.crypto.mac) {
            throw new Error('Key derivation failed - possibly wrong passphrase');
        }
        var decipher = cryptoUtils.crypto.createDecipheriv(json.crypto.cipher, derivedKey.slice(0, 16), new Buffer(json.crypto.cipherparams.iv, 'hex'));
        var seed = Buffer.concat([decipher.update(ciphertext), decipher.final()]);
        while (seed.length < 32) {
            var nullBuff = new Buffer([0x00]);
            seed = Buffer.concat([nullBuff, seed]);
        }
        return new Account(seed);
    }

};


module.exports = Account;

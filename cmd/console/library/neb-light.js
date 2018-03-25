require=(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){

"use strict";

var utils = require('./utils/utils.js');

/**
 * Admin API constructor.
 * Class encapsulate methods for admin APIs commands.
 * API documentation: {@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md}.
 * @constructor
 *
 * @param {Neb} neb - Instance of Neb library.
 *
 * @example
 * var admin = new Admin( new Neb() );
 * // or just
 * var admin = new Neb().admin;
 */
var Admin = function (neb) {
    this._setRequest(neb._request);
};

/**
 * @private
 * @param {Request} request - transport wrapper.
 */
Admin.prototype._setRequest = function (request) {
    this._request = request;
    this._path = '/admin';
};

/**
 * Method get info about nodes in Nebulas Network.
 *
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [nodeInfoObject]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#nodeinfo}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var info = admin.nodeInfo();
 * //async
 * admin.nodeInfo(function(info) {
 * //code
 * });
 */
Admin.prototype.nodeInfo = function () {
    var options = utils.argumentsToObject(['callback'], arguments);
    return this._sendRequest("get", "/nodeinfo", null, options.callback);
};

/**
 * Method get list of available addresses.
 *
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [accountsList]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#accounts}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var accounts = admin.accounts();
 * //async
 * admin.accounts(function(accounts) {
 * //code
 * });
 */
Admin.prototype.accounts = function () {
    var options = utils.argumentsToObject(['callback'], arguments);
    return this._sendRequest("get", "/accounts", null, options.callback);
};

/**
 * Method create a new account in Nebulas network with provided passphrase.
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#newaccount}
 *
 * @param {String} passphrase
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [address]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#newaccount}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var address = admin.newAccount("passphrase");
 * //async
 * admin.newAccount("passphrase", function(address) {
 * //code
 * });
 */
Admin.prototype.newAccount = function () {
    var options = utils.argumentsToObject(['passphrase', 'callback'], arguments);
    var params = { "passphrase": options.passphrase };
    return this._sendRequest("post", "/account/new", params, options.callback);
};

/**
 * Method unlock account with provided passphrase.
 * After the default unlock time, the account will be locked.
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#unlockaccount}
 *
 * @param {String} address
 * @param {String} passphrase
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [address]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#unlockaccount}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var isUnLocked = admin.unlockAccount("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf", "passphrase");
 * //async
 * admin.unlockAccount("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf", "passphrase", function(isUnLocked) {
 * //code
 * });
 */
Admin.prototype.unlockAccount = function () {
    var options = utils.argumentsToObject(['address', 'passphrase', 'callback'], arguments);
    var params = {
        "address": options.address,
        "passphrase": options.passphrase
    };
    return this._sendRequest("post", "/account/unlock", params, options.callback);
};

/**
 * Method lock account.
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#lockaccount}
 *
 * @param {String} address
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [address]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#lockaccount}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var isLocked = admin.lockAccount("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf");
 * //async
 * admin.lockAccount("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf", function(isLocked) {
 * //code
 * });
 */
Admin.prototype.lockAccount = function () {
    var options = utils.argumentsToObject(['address', 'callback'], arguments);
    var params = { "address": options.address };
    return this._sendRequest("post", "/account/lock", params, options.callback);
};

/**
 * Method wrap transaction sending functionality.
 * For more information about parameters, follow this link:
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#sendtransaction}
 *
 * @param {String} from
 * @param {String} to
 * @param {Number|Sting} value
 * @param {Number} nonce
 * @param {Number|String} gasPrice
 * @param {Number|String} gasLimit
 * @param {Object} [contract]
 * @param {String} [binary]
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [Transcation hash and contract address]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#sendtransaction}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var tx = admin.sendTransaction(
 *     "n1QsosVXKxiV3B4iDWNmxfN4VqpHn2TeUcn",
 *     "n1Lf5VcZQnzBc69iANxLTBqmojCeMFKowoM",
 *     "10",
 *     1,
 *     "1000000",
 *     "2000000"
 * );
 * //async
 * admin.sendTransaction(
 *     "n1QsosVXKxiV3B4iDWNmxfN4VqpHn2TeUcn",
 *     "n1Lf5VcZQnzBc69iANxLTBqmojCeMFKowoM",
 *     "10",
 *     1,
 *     "1000000",
 *     "2000000",
 *     null, null,
 *     function(tx) {
 *          //code
 *     }
 * );
 */
Admin.prototype.sendTransaction = function () {
    var options = utils.argumentsToObject(['from', 'to', 'value', 'nonce', 'gasPrice', 'gasLimit', 'contract', 'binary', 'callback'], arguments);
    var params = {
        "from": options.from,
        "to": options.to,
        "value": utils.toString(options.value),
        "nonce": options.nonce,
        "gasPrice": utils.toString(options.gasPrice),
        "gasLimit": utils.toString(options.gasLimit),
        "contract": options.contract,
        "binary": options.binary
    };
    return this._sendRequest("post", "/transaction", params, options.callback);
};

/**
 * Method sign hash.
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#signhash}
 *
 * @param {String} address
 * @param {string} string of hash bytes with base64 encode.
 * @param {uint32} alg
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [data]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#signhash}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var data = admin.SignHash("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf", "OGQ5NjllZWY2ZWNhZDNjMjlhM2E2MjkyODBlNjg2Y2YwYzNmNWQ1YTg2YWZmM2NhMTIwMjBjOTIzYWRjNmM5Mg==", 1);
 * //async
 * admin.SignHash("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf", "OGQ5NjllZWY2ZWNhZDNjMjlhM2E2MjkyODBlNjg2Y2YwYzNmNWQ1YTg2YWZmM2NhMTIwMjBjOTIzYWRjNmM5Mg==", 1, function(isLocked) {
 * //code
 * });
 */
Admin.prototype.signHash = function () {
    var options = utils.argumentsToObject(['address', 'hash', 'alg', 'callback'], arguments);
    var params = {
        "address": options.address,
        "hash": options.hash,
        "alg": options.alg
    };
    return this._sendRequest("post", "/sign/hash", params, options.callback);
};

/**
 * Method sign transaction with passphrase.
 * The transaction's from addrees must be unlock before sign call.
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#signtransactionwithpassphrase}
 *
 * @param {String} from
 * @param {String} to
 * @param {Number|Sting} value
 * @param {Number} nonce
 * @param {Number|String} gasPrice
 * @param {Number|String} gasLimit
 * @param {Object} [contract]
 * @param {String} [binary]
 * @param {String} passphrase
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [data]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#signtransactionwithpassphrase}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var data = admin.signTransactionWithPassphrase(
 *     "22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09",
 *     "5bed67f99cb3319e0c6f6a03548be3c8c52a8364464f886f",
 *     "10",
 *     1,
 *     "1000000",
 *     "2000000"
 *     null, null,
 *     "passphrase"
 * );
 * //async
 * admin.signTransactionWithPassphrase(
 *     "22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09",
 *     "5bed67f99cb3319e0c6f6a03548be3c8c52a8364464f886f",
 *     "10",
 *     1,
 *     "1000000",
 *     "2000000",
 *     null, null,
 *     "passphrase"
 *     function(data) {
 *          //code
 *     }
 * );
 */
Admin.prototype.signTransactionWithPassphrase = function () {
    var options = utils.argumentsToObject(['from', 'to', 'value', 'nonce', 'gasPrice', 'gasLimit', 'contract', 'binary', 'passphrase', 'callback'], arguments);
    var tx = {
        "from": options.from,
        "to": options.to,
        "value": utils.toString(options.value),
        "nonce": options.nonce,
        "gasPrice": utils.toString(options.gasPrice),
        "gasLimit": utils.toString(options.gasLimit),
        "contract": options.contract,
        "binary": options.binary
    };
    var params = {
        "transaction": tx,
        "passphrase": options.passphrase
    };
    return this._sendRequest("post", "/sign", params, options.callback);
};

/**
 * Method send transaction with passphrase.
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#sendtransactionwithpassphrase}
 *
 * @param {String} from
 * @param {String} to
 * @param {Number|Sting} value
 * @param {Number} nonce
 * @param {Number|String} gasPrice
 * @param {Number|String} gasLimit
 * @param {Object} [contract]
 * @param {String} [binary]
 * @param {String} [passphrase]
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [data]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#sendtransactionwithpassphrase}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var data = admin.sendTransactionWithPassphrase(
 *     "22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09",
 *     "5bed67f99cb3319e0c6f6a03548be3c8c52a8364464f886f",
 *     "10",
 *     1,
 *     "1000000",
 *     "2000000",
 *     null, null,
 *     "passphrase"
 * );
 * //async
 * admin.sendTransactionWithPassphrase
 *     "22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09",
 *     "5bed67f99cb3319e0c6f6a03548be3c8c52a8364464f886f",
 *     "10",
 *     1,
 *     "1000000",
 *     "2000000",
 *     null, null,
 *     "passphrase",
 *     function(data) {
 *          //code
 *     }
 * );
 */
Admin.prototype.sendTransactionWithPassphrase = function () {
    var options = utils.argumentsToObject(['from', 'to', 'value', 'nonce', 'gasPrice', 'gasLimit', 'contract', 'binary', 'passphrase', 'callback'], arguments);
    var tx = {
        "from": options.from,
        "to": options.to,
        "value": utils.toString(options.value),
        "nonce": options.nonce,
        "gasPrice": utils.toString(options.gasPrice),
        "gasLimit": utils.toString(options.gasLimit),
        "contract": options.contract,
        "binary": options.binary
    };
    var params = {
        "transaction": tx,
        "passphrase": options.passphrase
    };
    return this._sendRequest("post", "/transactionWithPassphrase", params, options.callback);
};

/**
 * Method start listen provided port. {@link https://github.com/nebulasio/go-nebulas/blob/1bd9bc9c9c6ca4fa0d515b620aa096f7e1c45088/neblet/neblet.go#L159}<br>
 * TODO: Add parameter to wiki documentation.
 *
 * @param {String} [callback] - Listen port.
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [isListenStrted]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#stopmining}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var isListenStrted = admin.startPprof('8080');
 * //async
 * admin.startPprof('8080', function(isListenStrted) {
 * //code
 * });
 */
Admin.prototype.startPprof = function () {
    var options = utils.argumentsToObject(['listen', 'callback'], arguments);
    var params = { "listen": options.listen };
    return this._sendRequest("post", "/pprof", params, options.callback);
};

/**
 * Method get config of node in Nebulas Network.
 *
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [config]{@link https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#getConfig}
 *
 * @example
 * var admin = new Neb().admin;
 * //sync
 * var info = admin.getConfig();
 * //async
 * admin.getConfig(function(info) {
 * //code
 * });
 */
Admin.prototype.getConfig = function () {
    var options = utils.argumentsToObject(['callback'], arguments);
    return this._sendRequest("get", "/getConfig", null, options.callback);
};

Admin.prototype._sendRequest = function (method, api, params, callback) {
    var action = this._path + api;
    if (typeof callback === "function") {
        return this._request.asyncRequest(method, action, params, callback);
    } else {
        return this._request.request(method, action, params);
    }
};

module.exports = Admin;

},{"./utils/utils.js":4}],2:[function(require,module,exports){

"use strict";

var utils = require('./utils/utils.js');

/**
 * User API constructor.
 * Class encapsulate methods for building distributed applications and services.
 * API documentation: {@link https://github.com/nebulasio/wiki/blob/master/rpc.md}.
 * @constructor
 *
 * @param {Neb} neb - Instance of Neb library.
 *
 * @example
 * var api = new API ( new Neb() );
 * // or just
 * var api = new Neb().api;
 */
var API = function (neb) {
    this._setRequest(neb._request);
};

/**
 * @private
 * @param {Request} request - transport wrapper.
 */
API.prototype._setRequest = function (request) {
    this._request = request;
    this._path = '/user';
};

/**
 * Method get state of Nebulas Network.
 *
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [NebStateObject]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#getnebstate}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var state = api.getNebState();
 * //async
 * api.getNebState(function(state) {
 * //code
 * });
 */
API.prototype.getNebState = function () {
    var options = utils.argumentsToObject(['callback'], arguments);
    return this._sendRequest("get", "/nebstate", null, options.callback);
};

/**
 * Method get latest irreversible block of Nebulas Network.
 *
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [NebStateObject]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#latestirreversibleblock}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var state = api.latestIrreversibleBlock();
 * //async
 * api.latestIrreversibleBlock(function(state) {
 * //code
 * });
 */
API.prototype.latestIrreversibleBlock = function () {
    var options = utils.argumentsToObject(['callback'], arguments);
    return this._sendRequest("get", "/lib", null, options.callback);
};

/**
 * Method return the state of the account. Balance and nonce.
 * For more information about parameters, follow this link:
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc.md#getaccountstate}
 *
 * @param {String} address
 * @param {String} height
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [accaountStateObject]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#getaccountstate}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var state = api.getAccountState("n1QsosVXKxiV3B4iDWNmxfN4VqpHn2TeUcn");
 * //async
 * api.getAccountState("n1QsosVXKxiV3B4iDWNmxfN4VqpHn2TeUcn", function(state) {
 * //code
 * });
 */
API.prototype.getAccountState = function () {
    var options = utils.argumentsToObject(['address', 'height', 'callback'], arguments);
    var params = { "address": options.address, "height": options.height };
    return this._sendRequest("post", "/accountstate", params, options.callback);
};

/**
 * Method wrap smart contract call functionality.
 * For more information about parameters, follow this link:
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc.md#call}
 *
 * @param {String} from
 * @param {String} to
 * @param {Number|Sting} value
 * @param {Number} nonce
 * @param {Number|String} gasPrice
 * @param {Number|String} gasLimit
 * @param {Object} contract
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [Transcation hash]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#call}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var tx = api.call(
 *     "n1QsosVXKxiV3B4iDWNmxfN4VqpHn2TeUcn",
 *     "n1Lf5VcZQnzBc69iANxLTBqmojCeMFKowoM",
 *     "0",
 *     3,
 *     "1000000",
 *     "2000000",
 *     "contract":{"function":"save","args":"[0]"}
 * );
 * //async
 * api.call(
 *     100,
 *     "n1QsosVXKxiV3B4iDWNmxfN4VqpHn2TeUcn",
 *     "n1Lf5VcZQnzBc69iANxLTBqmojCeMFKowoM",
 *     "0",
 *      3,
 *      "1000000",
 *      "2000000",
 *      "contract":{"function":"save","args":"[0]"},
 *      function(tx) {
 *          //code
 *      }
 * );
 */
API.prototype.call = function () {
    var options = utils.argumentsToObject(['from', 'to', 'value', 'nonce', 'gasPrice', 'gasLimit', 'contract', 'callback'], arguments);
    var params = {
        "from": options.from,
        "to": options.to,
        "value": utils.toString(options.value),
        "nonce": options.nonce,
        "gasPrice": utils.toString(options.gasPrice),
        "gasLimit": utils.toString(options.gasLimit),
        "contract": options.contract
    };
    return this._sendRequest("post", "/call", params, options.callback);
};

/**
 * Method wrap submit the signed transaction.
 * For more information about parameters, follow this link:
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc.md#sendrawtransaction}
 *
 * @param {Object} data
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [Transcation hash]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#sendrawtransaction}
 *
 * @example
 * var api = new Neb().api;
 * var tx = new Transaction(ChainID, from, to, transferValue, nonce, gasPrice, gasLimit);
 * tx.signTransaction();
 * //sync
 * var hash = api.sendRawTransaction( tx.toProtoString() );
 * //async
 * api.sendRawTransaction( tx.toProtoString(), function(hash) {
 * //code
 * });
 */
API.prototype.sendRawTransaction = function () {
    var options = utils.argumentsToObject(['data', 'callback'], arguments);
    var params = { "data": options.data };
    return this._sendRequest("post", "/rawtransaction", params, options.callback);
};

/**
 * Get block header info by the block hash.
 * For more information about parameters, follow this link:
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc.md#getblockbyhash}
 *
 * @param {String} hash
 * @param {Boolean} fullTransaction
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [Block]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#getblockbyhash}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var block = api.getBlockByHash("00000658397a90df6459b8e7e63ad3f4ce8f0a40b8803ff2f29c611b2e0190b8", true);
 * //async
 * api.getBlockByHash("00000658397a90df6459b8e7e63ad3f4ce8f0a40b8803ff2f29c611b2e0190b8", true,  function(block) {
 * //code
 * });
 */
API.prototype.getBlockByHash = function () {
    var options = utils.argumentsToObject(['hash', 'fullTransaction', 'callback'], arguments);
    var params = { "hash": options.hash, "fullTransaction": options.fullTransaction };
    return this._sendRequest("post", "/getBlockByHash", params, options.callback);
};

/**
 * Get block header info by the block height.
 * For more information about parameters, follow this link:
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc.md#getblockbyheight}
 *
 * @param {Number} height
 * @param {Boolean} fullTransaction
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [Block]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#getblockbyheight}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var block = api.getBlockByHeight(2, true);
 * //async
 * api.getBlockByHeight(2, true,  function(block) {
 * //code
 * });
 */
API.prototype.getBlockByHeight = function () {
    var options = utils.argumentsToObject(['height', 'fullTransaction', 'callback'], arguments);
    var params = { "height": options.height, "fullTransaction": options.fullTransaction };
    return this._sendRequest("post", "/getBlockByHeight", params, options.callback);
};

/**
 * Get transactionReceipt info by tansaction hash.
 * For more information about parameters, follow this link:
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc.md#gettransactionreceipt}
 *
 * @param {String} hash
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [TransactionReceipt]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#gettransactionreceipt}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var receipt = api.getTransactionReceipt("cc7133643a9ae90ec9fa222871b85349ccb6f04452b835851280285ed72b008c");
 * //async
 * api.getTransactionReceipt("cc7133643a9ae90ec9fa222871b85349ccb6f04452b835851280285ed72b008c", function(receipt) {
 * //code
 * });
 */
API.prototype.getTransactionReceipt = function () {
    var options = utils.argumentsToObject(['hash', 'callback'], arguments);
    var params = { "hash": options.hash };
    return this._sendRequest("post", "/getTransactionReceipt", params, options.callback);
};

/**
 * Return the subscribed events of transaction & block.
 * For more information about parameters, follow this link:
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc.md#subscribe}
 *
 * @param {Array|String} topic
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [eventData]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#subscribe}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var eventData = api.subscribe(["chain.linkBlock", "chain.pendingTransaction"]);
 * //async
 * api.subscribe(["chain.linkBlock", "chain.pendingTransaction"], function(eventData) {
 * //code
 * });
 */
API.prototype.subscribe = function () {
    var options = utils.argumentsToObject(['topic', 'callback'], arguments);
    var params = { "topic": options.topic };
    return this._sendRequest("post", "/subscribe", params, options.callback);
};

/**
 * Return current gasPrice.
 *
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [Gas Price]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#getgasprice}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var gasPrice = api.gasPrice();
 * //async
 * api.gasPrice(function(gasPrice) {
 * //code
 * });
 */
API.prototype.gasPrice = function () {
    var options = utils.argumentsToObject(['callback'], arguments);
    return this._sendRequest("get", "/getGasPrice", null, options.callback);
};

/**
 * Return the estimate gas of transaction.
 * For more information about parameters, follow this link:
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc.md#estimategas}
 *
 * @param {String} from
 * @param {String} to
 * @param {Number|Sting} value
 * @param {Number} nonce
 * @param {Number|String} gasPrice
 * @param {Number|String} gasLimit
 * @param {Object} [contract]
 * @param {String} [binary]
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [Gas]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#estimategas}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var gas = api.estimateGas(
 *     "n1QsosVXKxiV3B4iDWNmxfN4VqpHn2TeUcn",
 *     "n1Lf5VcZQnzBc69iANxLTBqmojCeMFKowoM",
 *     "10",
 *     1,
 *     "1000000",
 *     "2000000"
 * );
 * //async
 * api.estimateGas(
 *     "n1QsosVXKxiV3B4iDWNmxfN4VqpHn2TeUcn",
 *     "n1Lf5VcZQnzBc69iANxLTBqmojCeMFKowoM",
 *     "10",
 *     1,
 *     "1000000",
 *     "2000000",
 *     null, null,
 *     function(gas) {
 *          //code
 *     }
 * );
 */
API.prototype.estimateGas = function () {
    var options = utils.argumentsToObject(['from', 'to', 'value', 'nonce', 'gasPrice', 'gasLimit', 'contract', 'binary', 'callback'], arguments);
    var params = {
        "from": options.from,
        "to": options.to,
        "value": utils.toString(options.value),
        "nonce": options.nonce,
        "gasPrice": utils.toString(options.gasPrice),
        "gasLimit": utils.toString(options.gasLimit),
        "contract": options.contract,
        "binary": options.binary
    };
    return this._sendRequest("post", "/estimateGas", params, options.callback);
};

/**
 * Return the events list of transaction.
 * For more information about parameters, follow this link:
 * {@link https://github.com/nebulasio/wiki/blob/master/rpc.md#geteventsbyhash}
 *
 * @param {String} hash
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [Events]{@link https://github.com/nebulasio/wiki/blob/master/rpc.md#geteventsbyhash}
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var events = api.getEventsByHash("ec239d532249f84f158ef8ec9262e1d3d439709ebf4dd5f7c1036b26c6fe8073");
 * //async
 * api.getEventsByHash("ec239d532249f84f158ef8ec9262e1d3d439709ebf4dd5f7c1036b26c6fe8073", function(events) {
 * //code
 * });
 */
API.prototype.getEventsByHash = function () {
    var options = utils.argumentsToObject(['hash', 'callback'], arguments);
    var params = { "hash": options.hash };
    return this._sendRequest("post", "/getEventsByHash", params, options.callback);
};

/**
 * Method getter for dpos dynasty.{@link https://github.com/nebulasio/go-nebulas/blob/0c3439f9cedc539f64f64dd400878d2318cb215f/rpc/api_service.go#L596}<br>
 * TODO: Add parameter to wiki documentation.
 *
 * @param {Function} [callback] - Without callback return data synchronous.
 *
 * @return [delegatees]
 *
 * @example
 * var api = new Neb().api;
 * //sync
 * var delegatees = api.getDynasty();
 * //async
 * api.getDynasty(function(delegatees) {
 * //code
 * });
 */
API.prototype.getDynasty = function (height, callback) {
    var params = { "height": height };
    return this._sendRequest("post", "/dynasty", params, callback);
};

API.prototype._sendRequest = function (method, api, params, callback) {
    var action = this._path + api;
    if (typeof callback === "function") {
        return this._request.asyncRequest(method, action, params, callback);
    } else {
        return this._request.request(method, action, params);
    }
};

module.exports = API;

},{"./utils/utils.js":4}],3:[function(require,module,exports){

"use strict";

var BigNumber = require('bignumber.js');
var utils = require('./utils.js');

var unitMap = {
  'none': '0',
  'None': '0',
  'wei': '1',
  'Wei': '1',
  'kwei': '1000',
  'Kwei': '1000',
  'mwei': '1000000',
  'Mwei': '1000000',
  'gwei': '1000000000',
  'Gwei': '1000000000',
  'nas': '1000000000000000000',
  'NAS': '1000000000000000000'
};

var unitValue = function (unit) {
  unit = unit ? unit.toLowerCase() : 'nas';
  var unitValue = unitMap[unit];
  if (unitValue === undefined) {
    throw new Error('The unit undefined, please use the following units:' + JSON.stringify(unitMap, null, 2));
  }
  return new BigNumber(unitValue, 10);
};

var toBasic = function (number, unit) {
  return utils.toBigNumber(number).times(unitValue(unit));
};

var fromBasic = function (number, unit) {
  return utils.toBigNumber(number).dividedBy(unitValue(unit));
};

var nasToBasic = function (number) {
  return utils.toBigNumber(number).times(unitValue("nas"));
};

module.exports = {
  toBasic: toBasic,
  fromBasic: fromBasic,
  nasToBasic: nasToBasic
};

},{"./utils.js":4,"bignumber.js":"bignumber.js"}],4:[function(require,module,exports){

"use strict";

var BigNumber = require('bignumber.js');

var isNull = function (v) {
	return v === null || typeof v === "undefined";
};

var isBrowser = function () {
	return typeof window !== "undefined";
};

var isBigNumber = function (obj) {
	return obj instanceof BigNumber || obj && obj.constructor && obj.constructor.name === 'BigNumber';
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

var isNumber = function (object) {
	return typeof object === 'number';
};

var toBigNumber = function (number) {
	number = number || 0;
	if (isBigNumber(number)) {
		return number;
	}
	if (isString(number) && number.indexOf('0x') === 0) {
		return new BigNumber(number.replace('0x', ''), 16);
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

// Transform Array-like arguments object to common array.
var argumentsToArray = function (args) {
	var len = args.length,
	    resultArray = new Array(len);

	for (var i = 0; i < len; i += 1) {
		resultArray[i] = args[i];
	}
	return resultArray;
};

// Create object based on provided arrays
var zipArraysToObject = function (keysArr, valuesArr) {
	var resultObject = {};

	for (var i = 0; i < keysArr.length; i += 1) {
		resultObject[keysArr[i]] = valuesArr[i];
	}
	return resultObject;
};

// Function what make overall view for arguments.
// If arguments was provided separated by commas like "func(arg1 ,arg2)" we create
// ArgumentsObject and write keys from argsNames and value from args.
// in case wheare we provide args in object like "func({arg1: value})"
// we just return that object
var argumentsToObject = function (keys, args) {
	var ArgumentsObject = {};

	args = argumentsToArray(args);
	if (isObject(args[0])) {
		ArgumentsObject = args[0];
	} else {
		ArgumentsObject = zipArraysToObject(keys, args);
	}

	return ArgumentsObject;
};

module.exports = {
	isNull: isNull,
	isBrowser: isBrowser,
	isBigNumber: isBigNumber,
	isString: isString,
	isObject: isObject,
	isFunction: isFunction,
	isNumber: isNumber,
	toBigNumber: toBigNumber,
	toString: toString,
	argumentsToObject: argumentsToObject,
	zipArraysToObject: zipArraysToObject
};

},{"bignumber.js":"bignumber.js"}],5:[function(require,module,exports){

},{}],"bignumber.js":[function(require,module,exports){
'use strict';

module.exports = BigNumber; // jshint ignore:line

},{}],"neb":[function(require,module,exports){

"use strict";

var API = require("./api.js");
var Admin = require("./admin.js");

var Unit = require("./utils/unit.js");

/**
 * Neb API library constructor.
 * @constructor
 * @param {Request} request - transport wrapper.
 */
var Neb = function (request) {
	if (request) {
		this._request = request;
	}

	this.api = new API(this);
	this.admin = new Admin(this);
};

Neb.prototype.setRequest = function (request) {
	this._request = request;
	this.api._setRequest(request);
	this.admin._setRequest(request);
};

Neb.prototype.toBasic = Unit.toBasic;
Neb.prototype.fromBasic = Unit.fromBasic;
Neb.prototype.nasToBasic = Unit.nasToBasic;

module.exports = Neb;

},{"./admin.js":1,"./api.js":2,"./utils/unit.js":3}]},{},[])
//# sourceMappingURL=neb-light.js.map

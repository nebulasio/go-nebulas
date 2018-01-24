"use strict";

// var XMLHttpRequest = require("xmlhttprequest").XMLHttpRequest;
// var Neb = require("../dist/neb-node");
//
// var neb = new Neb();
//
// console.log(neb.api.accounts());
// var state = neb.api.getAccountState("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf");
// console.log(state);
// var result = neb.admin.unlockAccount("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf", "passphrase");
// console.log(result);
// result = neb.api.sendTransaction("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf", "22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09", neb.nasToBasic(5), parseInt(state.nonce)+1);
// console.log(result);
// state = neb.api.getAccountState("22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09");
// console.log(state);

var Account = require("../lib/account");

var account = Account.NewAccount();
console.log(account.getPrivateKeyString());
console.log(account.getPublicKeyString());
console.log(account.getAddressString());
console.log(Account.isValidAddress(account.getAddressString()));
var key = account.toKey("passphrase");
console.log(JSON.stringify(key));
console.log("********************");
var a1 = new Account();
a1 = a1.fromKey(key, "passphrase");
console.log(a1.getPrivateKeyString());

var Transaction = require("../lib/transaction");

var tx = new Transaction(100, account, account, "10", 1);
tx.signTransaction();
console.log("hash:" + tx.hash.toString("hex"));
console.log("sign:" + tx.sign.toString("hex"));
console.log(tx.toString());
var data = tx.toProtoString();
console.log(data);
tx.fromProto(data);
console.log(tx.toString());
console.log("address:"+tx.from.getAddressString());



var cryptoUtils = require("../lib/utils/crypto-utils");
console.log("black：" + cryptoUtils.sha3("").toString("hex"));
console.log("Hello, world：" + cryptoUtils.sha3("Hello, world").toString("hex"));

"use strict";

// var XMLHttpRequest = require("xmlhttprequest").XMLHttpRequest;
var Neb = require("../dist/neb-node");

var neb = new Neb();

console.log(neb.api.accounts());
var state = neb.api.getAccountState("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf");
console.log(state);
var result = neb.admin.unlockAccount("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf", "passphrase");
console.log(result);
result = neb.api.sendTransaction("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf", "22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09", neb.nasToBasic(5), parseInt(state.nonce)+1);
console.log(result);
state = neb.api.getAccountState("22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09");
console.log(state);
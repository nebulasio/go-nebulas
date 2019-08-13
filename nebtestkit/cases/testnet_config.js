'use strict';

var Wallet = require("nebulas");


var TestNet = function (env) {

    this.ChainId = "";
    this.SourceAccount = "";
    this.coinbase = "";
    this.apiEndPoint = "";

    if (env === 'testneb1') {

        this.ChainId = 1001;
        this.sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
        this.coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
        this.apiEndPoint = "http://13.57.120.136:8685";

    } else if (env === "testneb2") {

        this.ChainId = 1002;
        this.sourceAccount = new Wallet.Account("1d3fe06a53919e728315e2ccca41d4aa5b190845a79007797517e62dbc0df454");
        this.coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
        this.apiEndPoint = "http://34.205.26.12:8685";

    } else if (env === "testneb3") {

        this.ChainId = 1003;
        this.sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
        this.coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
    //    this.apiEndPoint = "http://35.182.205.40:8685";
        
        this.apiEndPoint = "http://13.57.120.136:8685";

    } else if (env === "testneb4") { //super node

        this.ChainId = 1004;
        this.sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
        this.coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
        this.apiEndPoint = "http://35.154.108.11:8685";

    } else if (env === "maintest") {

        this.ChainId = 2;
        this.sourceAccount = new Wallet.Account("d2319a8a63b1abcb0cc6d4183198e5d7b264d271f97edf0c76cfdb1f2631848c");
        this.coinbase = "n1dZZnqKGEkb1LHYsZRei1CH6DunTio1j1q";
        this.apiEndPoint = "https://mainnet.nebulas.io";

    } else if (env === "local") {

        this.ChainId = 100;
        this.sourceAccount = new Wallet.Account("1d3fe06a53919e728315e2ccca41d4aa5b190845a79007797517e62dbc0df454");
        this.coinbase = "n1XkoVVjswb5Gek3rRufqjKNpwrDdsnQ7Hq";
        this.apiEndPoint = "http://127.0.0.1:8685";

    } else if (env === "devnet") {

        this.ChainId = 1111;
        this.sourceAccount = new Wallet.Account("830ccbac2029b880eb07aa9a19c65ce6dad41702d409771eada791d6a6a83a1e");
        this.coinbase = "n1XkoVVjswb5Gek3rRufqjKNpwrDdsnQ7Hq";
        this.apiEndPoint = "http://47.92.203.173:9695";

    } else {
        console.log("====> could not found specified env: " + env + ", using local env instead");
        console.log("====> example: mocha cases/contract/xxx testneb2 -t 2000000");
        env = "local";
        // throw new Error("invalid env (" + env + ").");

        this.ChainId = 100;
        this.sourceAccount = new Wallet.Account("1d3fe06a53919e728315e2ccca41d4aa5b190845a79007797517e62dbc0df454");
        this.coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
        this.apiEndPoint = "http://127.0.0.1:8685";

    }
    console.log("\n=====> running: env:", env, " ChainId:", this.ChainId, " apiEndPoint:", this.apiEndPoint, " time:", new Date());

};

module.exports = TestNet;
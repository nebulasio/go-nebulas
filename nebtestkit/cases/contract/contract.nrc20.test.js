'use strict';

var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var sleep = require("system-sleep");
var HttpRequest = require("../../node-request");
var FS = require("fs");

var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var Neb = Wallet.Neb;
var neb = new Neb();

var ChainID;
var source, deploy, from, fromState, contractAddr;

var coinbase, coinState;
var testCases = new Array();
var caseIndex = 0;

// mocha cases/contract/xxx testneb1 -t 200000
var args = process.argv.splice(2);
var env = args[1];
if (env !== "local" && env !== "testneb1" && env !== "testneb2") {
    env = "local";
}
console.log("env:", env);

if (env == 'local'){
    neb.setRequest(new HttpRequest("http://127.0.0.1:8685"));//https://testnet.nebulas.io
    ChainID = 100;
    source = new Wallet.Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
    coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
}else if(env == 'testneb1'){
    neb.setRequest(new HttpRequest("http://35.182.48.19:8685"));
    ChainID = 1001;
    source = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
}else if(env == "testneb2"){
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
    ChainID = 1002;
    source = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
}else{
    console.log("please input correct env local testneb1 testneb2");
    return;
}

var lastnonce = 0;

// deploy = new Wallet.Account("823e8a73257beb9f8ddc5c10ec32b886199278d75371a9c6fdd33f8f4ea5b792");
// contractAddr = "249596e82b086f76df1310d36b475ab33f5595fcec7d61aa";

function prepareContractCall(testCase, done) {
    neb.api.getAccountState(source.getAddressString()).then(function (resp) {
        console.log("source account state:" + JSON.stringify(resp));
        lastnonce = parseInt(resp.nonce);

        var accounts = new Array();
        var values = new Array();
        if (Wallet.Utils.isNull(contractAddr)) {
            deploy = Wallet.Account.NewAccount();
            accounts.push(deploy);
            values.push(neb.nasToBasic(1));
        }

        if (typeof testCase.testInput.from !== "undefined") {
            accounts.push(testCase.testInput.from);
            values.push(neb.nasToBasic(1));
        }

        if (typeof testCase.testInput.to !== "undefined") {
            accounts.push(testCase.testInput.to);
            values.push(neb.nasToBasic(1));
        }

        if (accounts.length > 0) {
            cliamTokens(accounts, values, function () {
                if (Wallet.Utils.isNull(contractAddr)) {
                    deployContract(done);
                } else {
                    done();
                }
            });
        } else {
            done();
        }

    });
}

function cliamTokens(accounts, values, done) {
    for (var i = 0; i < accounts.length; i++) {
        // console.log("acc:"+accounts[i].getAddressString()+"value:"+values[i]);
        sendTransaction(source, accounts[i], values[i], ++lastnonce);
        sleep(30);
    }
    checkCliamTokens(done);
}

function sendTransaction(from, address, value, nonce) {
    var transaction = new Wallet.Transaction(ChainID, from, address, value, nonce, "1000000", "2000000");
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();
    // console.log("send transaction:", transaction.toString());
    neb.api.sendRawTransaction(rawTx).then(function (resp) {
        console.log("send raw transaction resp:" + JSON.stringify(resp));
    });
}

function checkCliamTokens(done) {
    var intervalAccount = setInterval(function () {
        neb.api.getAccountState(source.getAddressString()).then(function (resp) {
            // console.log("master accountState resp:" + JSON.stringify(resp));
            var nonce = parseInt(resp.nonce);
            console.log("check cliam tokens nonce:", lastnonce);

            if (lastnonce <= nonce){
                console.log("cliam tokens success");
                clearInterval(intervalAccount);

                done();
            }
        });
    }, 2000);
}

function deployContract(done){

    // create contract
    var source = FS.readFileSync("../nf/nvm/test/NRC20.js", "utf-8");
    var contract = {
        "source": source,
        "sourceType": "js",
        "args": "[\"StandardToken\", \"NRC\", 18, \"1000000000\"]"
    };

    var transaction = new Wallet.Transaction(ChainID, deploy, deploy, "0", 1, "10000000", "20000000000", contract);
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();

    // console.log("contract:" + rawTx);

    neb.api.sendRawTransaction(rawTx).then(function (resp) {
        console.log("deploy contract:" + JSON.stringify(resp));

        checkTransaction(resp.txhash, done);
    });
}

function checkTransaction(txhash, done){

    var retry = 0;
    var maxRetry = 20;

    // contract status and get contract_address
    var interval = setInterval(function () {
        // console.log("getTransactionReceipt hash:"+txhash);
        neb.api.getTransactionReceipt(txhash).then(function (resp) {
            retry++;

            console.log("check transaction status:" + resp.status);

            if(resp.status && resp.status === 1) {
                clearInterval(interval);

                if (resp.contract_address) {
                    console.log("deploy private key:" + deploy.getPrivateKeyString());
                    console.log("deploy address:" + deploy.getAddressString());
                    console.log("deploy contract address:" + resp.contract_address);
                    // console.log("deploy receipt:" + JSON.stringify(resp));

                    contractAddr = resp.contract_address;

                    // checkNRCBalance(resp.from, resp.contract_address);
                }

                done(resp);
            } else if (resp.status && resp.status === 2) {
                if (retry > maxRetry) {
                    console.log("check transaction time out");
                    clearInterval(interval);
                    done(resp);
                }
            } else {
                clearInterval(interval);
                console.log("transaction execution failed");
                done(resp);
            }
        }).catch(function (err) {
            retry++;
            console.log("check transaction not found retry");
            if (retry > maxRetry) {
                console.log(JSON.stringify(err.error));
                clearInterval(interval);
                done(err);
            }
        });

    }, 2000);
}

function testCall(testInput, testExpect, done) {
    var contract = {
        "function": testInput.function,
        "args": testInput.args
    };
    var from = Wallet.Account.NewAccount();
    neb.api.call(from.getAddressString(), contractAddr, "0", 1, "1000000", "2000000", contract).then(function (resp) {
        var result = JSON.parse(resp.result);
        console.log("result:", result);
        expect(result).to.equal(testExpect.result);
        done();
    }).catch(function (err) {
        if (testExpect.exeFailed) {
            console.log("call failed:", err.message);
            done();
        } else {
            done(err);
        }
    });
}

function testTransfer(testInput, testExpect, done) {
    var from = (Wallet.Utils.isNull(testInput.from)) ? deploy : testInput.from;
    var to = Wallet.Account.NewAccount();
    var fromBalance, toBalance;

    balanceOfNRC20(from.getAddressString()).then(function(resp) {
        fromBalance = JSON.parse(resp.result);
        console.log("from balance:", fromBalance);

        return balanceOfNRC20(to.getAddressString());
    }).then(function (resp) {
        toBalance = JSON.parse(resp.result);
        console.log("to balance:", toBalance);

        return neb.api.getAccountState(from.getAddressString());
    }).then(function (resp) {
        console.log("from state:", JSON.stringify(resp));

        var args = testInput.args;
        if (!Wallet.Utils.isNull(testInput.transferValue)) {
            if (testInput.transferValue === "from.balance") {
                testInput.transferValue = fromBalance;
            }
            args = "[\""+ to.getAddressString() +"\", \""+ testInput.transferValue +"\"]";
        }

        var contract = {
            "function": "transfer",
            "args": args
        };
        var tx = new Wallet.Transaction(ChainID, from, contractAddr, "0", parseInt(resp.nonce) + 1, "1000000", "2000000", contract);
        tx.signTransaction();

        console.log("raw tx:", tx.toString());
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function (resp) {
        console.log("send raw tx:", resp);
        checkTransaction(resp.txhash, function (receipt) {
            var resetContract = false;
            try {
                if (testExpect.status === 1) {
                    expect(receipt).to.be.have.property('status').equal(testExpect.status);
                } else {
                    expect(receipt).to.not.have.property('status');
                }

                balanceOfNRC20(from.getAddressString()).then(function (resp) {
                    var balance = JSON.parse(resp.result);
                    console.log("after from balance:", balance);

                    if (testExpect.status === 1) {
                        var balanceNumber = new BigNumber(fromBalance).sub(testInput.transferValue);
                        expect(balanceNumber.toString(10)).to.equal(balance);
                    } else {
                        expect(balance).to.equal(fromBalance);
                    }

                    if (balance === "0") {
                        resetContract = true;
                    }

                    return balanceOfNRC20(to.getAddressString());
                }).then(function (resp) {
                    var balance = JSON.parse(resp.result);
                    console.log("after to balance:", balance);

                    if (testExpect.status === 1) {
                        var balanceNumber = new BigNumber(toBalance).plus(testInput.transferValue);
                        expect(balanceNumber.toString(10)).to.equal(balance);
                    } else {
                        expect(toBalance).to.equal(balance);
                    }

                    return neb.api.getEventsByHash(receipt.hash);
                }).then(function (events) {
                    // console.log("tx events:", events);
                    for (var i = 0; i < events.events.length; i++) {
                        var event = events.events[i];
                        console.log("tx event:", event);
                        if (event.topic == "chain.transactionResult") {
                            var result = JSON.parse(event.data);
                            expect(result.status).to.equal(testExpect.status);
                        }
                    }
                    if (resetContract) {
                        contractAddr = null;
                    }
                    done();
                }).catch(function (err) {
                    if (resetContract) {
                        contractAddr = null;
                    }
                    done(err);
                })
            } catch (err) {
                if (resetContract) {
                    contractAddr = null;
                }
                done(err);
            }
        });
    }).catch(function(err) {
        done(err);
    });
}

function testApprove(testInput, testExpect, done) {
    var from = (Wallet.Utils.isNull(testInput.from)) ? deploy : testInput.from;
    var to = Wallet.Account.NewAccount();
    var fromAllowance, fromBalance, fromState;

    allowanceOfNRC20(from.getAddressString(), to.getAddressString()).then(function (resp) {
        fromAllowance = JSON.parse(resp.result);
        console.log("allowance:", fromAllowance);

        return balanceOfNRC20(from.getAddressString());
    }).then(function (resp) {
        fromBalance = JSON.parse(resp.result);
        console.log("balance:", fromBalance);

        return neb.api.getAccountState(from.getAddressString());
    }).then(function (resp) {
        fromState = resp;
        console.log("from state:", resp);

        var args = testInput.args;
        if (!Wallet.Utils.isNull(testInput.approveValue)) {
            if (testInput.approveValue === "from.balance") {
                testInput.approveValue = fromBalance;
            }
            var currentValue = fromAllowance;
            if (!Wallet.Utils.isNull(testInput.currentValue)) {
                currentValue = testInput.currentValue;
            }
            args = "[\""+ to.getAddressString() +"\", \""+ currentValue +"\", \""+ testInput.approveValue +"\"]";
        }

        var contract = {
            "function": "approve",
            "args": args
        };
        var tx = new Wallet.Transaction(ChainID, from, contractAddr, "0", parseInt(resp.nonce) + 1, "1000000", "2000000", contract);
        tx.signTransaction();

        console.log("raw tx:", tx.toString());
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function (resp) {
        console.log("send raw tx:", resp);

        checkTransaction(resp.txhash, function (receipt) {
            try {
                if (testExpect.status === 1) {
                    expect(receipt).to.be.have.property('status').equal(testExpect.status);
                } else {
                    expect(receipt).to.not.have.property('status');
                }

                balanceOfNRC20(from.getAddressString()).then(function (resp) {
                    var balance = JSON.parse(resp.result);
                    console.log("after from balance:", balance);
                    expect(balance).to.equal(fromBalance);

                    return allowanceOfNRC20(from.getAddressString(), to.getAddressString());
                }).then(function (resp) {
                    var allownance = JSON.parse(resp.result);
                    console.log("after from allownance:", allownance);
                    if (testExpect.status === 1) {
                        expect(allownance).to.equal(testInput.approveValue);
                    }

                    return neb.api.getEventsByHash(receipt.hash);
                }).then(function (events) {
                    // console.log("tx events:", events);
                    for (var i = 0; i < events.events.length; i++) {
                        var event = events.events[i];
                        console.log("tx event:", event);
                        if (event.topic == "chain.transactionResult") {
                            var result = JSON.parse(event.data);
                            expect(result.status).to.equal(testExpect.status);
                        }
                    }
                    done();
                }).catch(function (err) {
                    done(err);
                })
            } catch (err) {
                done(err);
            }
        });

    }).catch(function (err) {
        done(err);
    });
}

function testTransferFrom(testInput, testExpect, done) {
    var from = (Wallet.Utils.isNull(testInput.from)) ? deploy : testInput.from;
    var to = (Wallet.Utils.isNull(testInput.to)) ? Wallet.Account.NewAccount() : testInput.to;
    var deployAllowance, deployBalance, deployState, fromBalance, toBalance;

    allowanceOfNRC20(deploy.getAddressString(), from.getAddressString()).then(function (resp) {
        deployAllowance = JSON.parse(resp.result);
        console.log("deploy allowance:", deployAllowance);

        return balanceOfNRC20(deploy.getAddressString());
    }).then(function (resp) {
        deployBalance = JSON.parse(resp.result);
        console.log("deploy balance:", deployBalance);

        return balanceOfNRC20(from.getAddressString());
    }).then(function (resp) {
        fromBalance = JSON.parse(resp.result);
        console.log("from balance:", fromBalance);

        return balanceOfNRC20(to.getAddressString());
    }).then(function (resp) {
        toBalance = JSON.parse(resp.result);
        console.log("to balance:", toBalance);

        return neb.api.getAccountState(deploy.getAddressString());
    }).then(function (resp) {
        deployState = resp;
        console.log("deploy state:", resp);

        return neb.api.getAccountState(from.getAddressString());
    }).then(function (resp) {
        var fromState = resp;
        console.log("from state:", resp);

        approveNRC20(testInput, deployState, from, deployAllowance, function (resp) {
            if (!Wallet.Utils.isNull(resp)) {
                if (resp instanceof Error) {
                    done(resp);
                }
                if (!(resp.status && resp.status === 1)) {
                    done(new Error("approve failed"));
                }
            }

            allowanceOfNRC20(deploy.getAddressString(), from.getAddressString()).then(function (resp) {
                deployAllowance = JSON.parse(resp.result);
                console.log("deploy allowance:", deployAllowance);

                var args = testInput.args;
                if (!Wallet.Utils.isNull(testInput.transferValue)) {
                    args = "[\""+ deploy.getAddressString() +"\", \""+ to.getAddressString() +"\", \""+ testInput.transferValue +"\"]";
                }

                var contract = {
                    "function": "transferFrom",
                    "args": args
                };
                var tx = new Wallet.Transaction(ChainID, from, contractAddr, "0", parseInt(fromState.nonce) + 1, "1000000", "2000000", contract);
                tx.signTransaction();

                console.log("raw tx:", tx.toString());
                return neb.api.sendRawTransaction(tx.toProtoString());
            }).then(function (resp) {
                console.log("send raw tx:", resp);

                checkTransaction(resp.txhash, function (receipt) {
                    try {
                        if (testExpect.status === 1) {
                            expect(receipt).to.be.have.property('status').equal(testExpect.status);
                        } else {
                            expect(receipt).to.not.have.property('status');
                        }

                        balanceOfNRC20(deploy.getAddressString()).then(function (resp) {
                            var balance = JSON.parse(resp.result);
                            console.log("after deploy balance:", balance);


                            if (testExpect.status === 1) {
                                var balanceNumber = new BigNumber(deployBalance).sub(testInput.transferValue);
                                expect(balanceNumber.toString(10)).to.equal(balance);
                            } else {
                                expect(balance).to.equal(deployBalance);
                            }

                            return balanceOfNRC20(from.getAddressString());
                        }).then(function (resp) {
                            var balance = JSON.parse(resp.result);
                            console.log("after from balance:", balance);
                            expect(balance).to.equal(fromBalance);

                            return allowanceOfNRC20(deploy.getAddressString(), from.getAddressString());
                        }).then(function (resp) {
                            var allownance = JSON.parse(resp.result);
                            console.log("after deploy allownance:", allownance);
                            if (testExpect.status === 1) {
                                var allownanceNumber = new BigNumber(deployAllowance).sub(testInput.transferValue);
                                expect(allownanceNumber.toString(10)).to.equal(allownance);
                            } else {
                                expect(deployAllowance).to.equal(allownance);
                            }

                            return balanceOfNRC20(to.getAddressString());
                        }).then(function (resp) {
                            var balance = JSON.parse(resp.result);
                            console.log("after to balance:", balance);

                            if (testExpect.status === 1) {
                                var balanceNumber = new BigNumber(toBalance).plus(testInput.transferValue);
                                expect(balanceNumber.toString(10)).to.equal(balance);
                            } else {
                                expect(toBalance).to.equal(balance);
                            }

                            return neb.api.getEventsByHash(receipt.hash);
                        }).then(function (events) {
                            // console.log("tx events:", events);
                            for (var i = 0; i < events.events.length; i++) {
                                var event = events.events[i];
                                console.log("tx event:", event);
                                if (event.topic == "chain.transactionResult") {
                                    var result = JSON.parse(event.data);
                                    expect(result.status).to.equal(testExpect.status);
                                }
                            }
                            done();
                        }).catch(function (err) {
                            done(err);
                        })
                    } catch (err) {
                        done(err);
                    }
                });
            });
        });
    }).catch(function (err) {
        done(err);
    });
}

function approveNRC20(testInput, deployState, from, currentValue, done) {
    if (!Wallet.Utils.isNull(testInput.approveValue)) {
        var approveValue = testInput.approveValue;
        var args = "[\""+ from.getAddressString() +"\", \""+ currentValue +"\", \""+ approveValue +"\"]";

        var contract = {
            "function": "approve",
            "args": args
        };
        var tx = new Wallet.Transaction(ChainID, deploy, contractAddr, "0", parseInt(deployState.nonce) + 1, "1000000", "2000000", contract);
        tx.signTransaction();
        // console.log("approve tx:", tx.toString());
        neb.api.sendRawTransaction(tx.toProtoString()).then(function (resp) {
            console.log("approve tx:", resp);

            checkTransaction(resp.txhash, done);
        }).catch(function (err) {
            done(err);
        });
    } else {
        done();
    }
}

function balanceOfNRC20(address) {
    var contract = {
        "function": "balanceOf",
        "args": "[\"" + address + "\"]"
    };
    return neb.api.call(address, contractAddr, "0", 1, "1000000", "200000", contract)
}

function allowanceOfNRC20(owner, spender) {
    var contract = {
        "function": "allowance",
        "args": "[\"" + owner + "\", \""+ spender +"\"]"
    };
    return neb.api.call(owner, contractAddr, "0", 1, "1000000", "2000000", contract)
}

var testCase = {
    "name": "1. name",
    "testInput": {
        isCall: true,
        function: "name",
        args: ""
    },
    "testExpect": {
        result: "StandardToken"
    }
};
testCases.push(testCase);


var testCase = {
    "name": "2. symbol",
    "testInput": {
        isCall: true,
        function: "symbol",
        args: ""
    },
    "testExpect": {
        result: "NRC"
    }
};
testCases.push(testCase);

testCase = {
    "name": "3. decimals",
    "testInput": {
        isCall: true,
        function: "decimals",
        args: ""
    },
    "testExpect": {
        result: 18
    }
};
testCases.push(testCase);

testCase = {
    "name": "4. totalSupply",
    "testInput": {
        isCall: true,
        function: "totalSupply",
        args: ""
    },
    "testExpect": {
        result: "1000000000000000000000000000"
    }
};
testCases.push(testCase);

testCase = {
    "name": "5. balanceOf args empty",
    "testInput": {
        isCall: true,
        function: "balanceOf",
        args: ""
    },
    "testExpect": {
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "6. balanceOf args format err",
    "testInput": {
        isCall: true,
        function: "balanceOf",
        args: "[1"
    },
    "testExpect": {
        exeFailed: true,
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "7. balanceOf args arr",
    "testInput": {
        isCall: true,
        function: "balanceOf",
        args: "[1]"
    },
    "testExpect": {
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "8. balanceOf address no balance",
    "testInput": {
        isCall: true,
        function: "balanceOf",
        args: "[\"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\"]"
    },
    "testExpect": {
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "9. balanceOf address have balance",
    "testInput": {
        isCall: true,
        function: "balanceOf",
        args: ""
    },
    "testExpect": {
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "10. allowance args empty",
    "testInput": {
        isCall: true,
        function: "allowance",
        args: ""
    },
    "testExpect": {
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "11. allowance args less",
    "testInput": {
        isCall: true,
        function: "allowance",
        args: "[\"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\"]"
    },
    "testExpect": {
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "12. allowance args format err",
    "testInput": {
        isCall: true,
        function: "allowance",
        args: "[\"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\""
    },
    "testExpect": {
        exeFailed: true,
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "13. allowance args err",
    "testInput": {
        isCall: true,
        function: "allowance",
        args: "[\"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\", 1]"
    },
    "testExpect": {
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "14. allowance args no allowance",
    "testInput": {
        isCall: true,
        function: "allowance",
        args: "[\"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\",\"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\"]"
    },
    "testExpect": {
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "15. allowance correct",
    "testInput": {
        isCall: true,
        function: "allowance",
        args: ""
    },
    "testExpect": {
        result: "0"
    }
};
testCases.push(testCase);

testCase = {
    "name": "16. transfer args empty",
    "testInput": {
        isTransfer: true,
        function: "transfer",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "17. transfer args less",
    "testInput": {
        isTransfer: true,
        function: "transfer",
        args: "[\"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\"]"
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "18. transfer args err",
    "testInput": {
        isTransfer: true,
        function: "transfer",
        args: "[0]"
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "19. transfer value = 0",
    "testInput": {
        isTransfer: true,
        transferValue: "0",
        function: "transfer",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

testCase = {
    "name": "20. transfer value < balance ",
    "testInput": {
        isTransfer: true,
        transferValue: "1",
        function: "transfer",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

testCase = {
    "name": "21. transfer value = balance ",
    "testInput": {
        isTransfer: true,
        transferValue: "from.balance",
        function: "transfer",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

testCase = {
    "name": "22. transfer value > balance ",
    "testInput": {
        isTransfer: true,
        transferValue: "100000000000000000000000000000000",
        function: "transfer",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "23. transfer balance = 0 & value = 0",
    "testInput": {
        isTransfer: true,
        from: Wallet.Account.NewAccount(),
        transferValue: "0",
        function: "transfer",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

testCase = {
    "name": "24. approve args empty",
    "testInput": {
        isApprove: true,
        function: "approve",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "25. approve args less",
    "testInput": {
        isApprove: true,
        function: "approve",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "26. approve args err",
    "testInput": {
        isApprove: true,
        function: "approve",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "27. approve balance = 0 & value = 0",
    "testInput": {
        isApprove: true,
        from: Wallet.Account.NewAccount(),
        function: "approve",
        approveValue: "0",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

testCase = {
    "name": "28. approve balance < value",
    "testInput": {
        isApprove: true,
        approveValue: "1000000000000000000000000000000",
        function: "approve",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "29. approve balance > value",
    "testInput": {
        isApprove: true,
        approveValue: "1",
        function: "approve",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

testCase = {
    "name": "30. approve balance = value",
    "testInput": {
        isApprove: true,
        approveValue: "from.balance",
        function: "approve",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

testCase = {
    "name": "31. approve value < 0",
    "testInput": {
        isApprove: true,
        approveValue: "-1",
        function: "approve",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

testCase = {
    "name": "32. approve currentValue  not correct",
    "testInput": {
        isApprove: true,
        currentValue: "123123",
        approveValue: "1",
        function: "approve",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "33. transferFrom args empty",
    "testInput": {
        isTransferFrom: true,
        from: Wallet.Account.NewAccount(),
        function: "transferFrom",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "34. transferFrom args less",
    "testInput": {
        isTransferFrom: true,
        from: Wallet.Account.NewAccount(),
        function: "transferFrom",
        args: "[\"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\"]"
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "35. transferFrom args err",
    "testInput": {
        isTransferFrom: true,
        from: Wallet.Account.NewAccount(),
        function: "transferFrom",
        args: "[\"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\", 1]"
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "36. transferFrom no approve",
    "testInput": {
        isTransferFrom: true,
        from: Wallet.Account.NewAccount(),
        transferValue: "1",
        function: "transferFrom",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "37. transferFrom approve < value",
    "testInput": {
        isTransferFrom: true,
        from: Wallet.Account.NewAccount(),
        approveValue: "10000",
        transferValue: "10000000000000000000000000000",
        function: "transferFrom",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);

testCase = {
    "name": "38. transferFrom approve > value",
    "testInput": {
        isTransferFrom: true,
        from: Wallet.Account.NewAccount(),
        approveValue: "10",
        transferValue: "1",
        function: "transferFrom",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

testCase = {
    "name": "39. transferFrom approve = value",
    "testInput": {
        isTransferFrom: true,
        from: Wallet.Account.NewAccount(),
        approveValue: "1",
        transferValue: "1",
        function: "transferFrom",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

testCase = {
    "name": "40. transferFrom approve = value = 0",
    "testInput": {
        isTransferFrom: true,
        from: Wallet.Account.NewAccount(),
        approveValue: "0",
        transferValue: "0",
        function: "transferFrom",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);

describe('contract call test', function () {
    // beforeEach(function (done) {
    //     prepareContractCall(function (result) {
    //         if (result instanceof Error) {
    //             done(result);
    //         } else {
    //             done();
    //         }
    //     });
    // });

    // var testCase = testCases[16];
    // it(testCase.name, function (done) {
    //     prepareContractCall(testCase, function (err) {
    //         if (err instanceof Error) {
    //             done(err);
    //         } else {
    //             if (testCase.testInput.isCall) {
    //                 testCall(testCase.testInput, testCase.testExpect, done);
    //             } else if (testCase.testInput.isTransfer) {
    //                 testTransfer(testCase.testInput, testCase.testExpect, done);
    //             } else if (testCase.testInput.isApprove) {
    //                 testApprove(testCase.testInput, testCase.testExpect, done);
    //             } else if (testCase.testInput.isTransferFrom) {
    //                 testTransferFrom(testCase.testInput, testCase.testExpect, done);
    //             }
    //         }
    //     });
    // });

    for (var i = 0; i < testCases.length; i++) {

        it(testCases[i].name, function (done) {
            var testCase = testCases[caseIndex];
            prepareContractCall(testCase, function (err) {
                if (err instanceof Error) {
                    done(err);
                } else {
                    if (testCase.testInput.isCall) {
                        testCall(testCase.testInput, testCase.testExpect, done);
                    } else if (testCase.testInput.isTransfer) {
                        testTransfer(testCase.testInput, testCase.testExpect, done);
                    } else if (testCase.testInput.isApprove) {
                        testApprove(testCase.testInput, testCase.testExpect, done);
                    } else if (testCase.testInput.isTransferFrom) {
                        testTransferFrom(testCase.testInput, testCase.testExpect, done);
                    }
                }
            });
        });
    }

    afterEach(function () {
        caseIndex++;
    });
});

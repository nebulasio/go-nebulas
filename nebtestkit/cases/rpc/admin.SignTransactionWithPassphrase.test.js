'use strict';

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet = require('nebulas');

var coinbase,
    sourceAccount,
    chain_id,
    server_address;

var env = process.env.NET || 'local';
var env = 'local';
if (env === 'testneb1') {
  chain_id = 1001;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "35.182.48.19:8684";

} else if (env === "testneb2") {
  chain_id = 1002;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "34.205.26.12:8684";

} else if (env === "testneb3") {
  chain_id = 1003;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "35.177.214.138:8684";

} else if (env === "testneb4") { //super node
  chain_id = 1004;
  sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
  coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
  server_address = "35.154.108.11:8684";
} else if (env === "testneb4_normalnode"){
  chain_id = 1004;
  sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
  coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
  server_address = "18.197.107.228:8684";
} else if (env === "local") {
  chain_id = 100;
  sourceAccount = new Wallet.Account("d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5");
  coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
  server_address = "127.0.0.1:8684";

} else if (env === "maintest"){
  chain_id = 2;
  sourceAccount = new Wallet.Account("d2319a8a63b1abcb0cc6d4183198e5d7b264d271f97edf0c76cfdb1f2631848c");
  coinbase = "n1dZZnqKGEkb1LHYsZRei1CH6DunTio1j1q";
  server_address = "54.149.15.132:8684";
} else {
  throw new Error("invalid env (" + env + ").");
}

var client,
    address;

function testSignTransaction(testInput, testExpect, done) {
    client.SignTransactionWithPassphrase(testInput, (err, resp) => {
        try {
            expect(!!err).to.equal(testExpect.hasError);

            if (err) {
                console.log(JSON.stringify(err));
                expect(err).have.property('details').equal(testExpect.errorMsg);
            } else {
                console.log(JSON.stringify(resp));
                expect(resp).to.have.property('data');
            }
            done();
        } catch (err) {
            done(err);
        }
    });
}

describe("rpc: SignTransaction with passphrase", () => {
    before((done) => {
        client = rpc_client.new_client(server_address, 'AdminService');

        try {
            client.NewAccount({passphrase: "passphrase"}, (err, resp) => {
                expect(!!err).to.be.false;
                expect(resp).to.have.property('address');
                address = resp.address;
                console.log("create new account: " + address);
                done();
            });
        } catch(err) {
            done(err)
        }
    });

    it('1. normal', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: false,
            errorMsg: "",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('2.  `to` illegal', done => {
        var testInput = {
             transaction: {
                from: address,
                to: "faaaa",
                value: "",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address: invalid address format",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('3.  `to` empty', done => {
        var testInput = {
            transaction: {
                from: address,
                to: "",
                value: "",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address: invalid address format",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('4.  `value` empty', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid value",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('5.  `value` alpha', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "abc",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid value",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('6.  `value` neg number', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "-123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid value",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    // it('7.  `nonce` alpha', done => {//nonce can not be string
    //     var testInput = {
    //         transaction: {
    //             from: address,
    //             to: coinbase,
    //             value: '$',
    //             nonce: '', 
    //             gas_price: "1000000",
    //             gas_limit: "1000000",
    //         },
    //         passphrase: 'passphrase'
    //     }

    //     var testExpect = {
    //         hasError: false,

    //     }

    //     testSignTransaction(testInput, testExpect, done)
    // });

    it('8.  `nonce` neg number', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: -100, 
                gas_price: "1000000",
                gas_limit: "1000000",
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: false,

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('9.  `gas_price` empty', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasPrice",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('10.  `gas_price` alpha', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "abcxz",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasPrice",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('11.  `gas_price` negative', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "-10000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasPrice",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('12.  `gas_limit` empty', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasLimit",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('13.  `gas_limit` alpha', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "aaz",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasLimit",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('14.  `gas_limit` negative', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "-10000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasLimit",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('15. `contract` empty', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid contract",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('16. `contract`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {
                    "function": "save",
                }
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid contract",

        }

        testSignTransaction(testInput, testExpect, done)
    });

it('17. `invalid type`', done => {
    var testInput = {
        transaction: {
            from: address,
            to: coinbase,
            value: "123",
            nonce: "10000",
            gas_price: "1000000",
            gas_limit: "1000000",
            contract: {
                "function": "save",
            },
            type: "invalid"
        },
        passphrase: 'passphrase'
    }

    var testExpect = {
        hasError: true,
        errorMsg: "invalid transaction data payload type",

    }

    testSignTransaction(testInput, testExpect, done)
});

it('18. `binary type`', done => {
    var testInput = {
        transaction: {
            from: address,
            to: coinbase,
            value: "123",
            nonce: "10000",
            gas_price: "1000000",
            gas_limit: "1000000",
            type: "binary"
        },
        passphrase: 'passphrase'
    }

    var testExpect = {
        hasError: false,
        errorMsg: "",

    }

    testSignTransaction(testInput, testExpect, done)
});

it('19. `deploy type`', done => {
    var testInput = {
        transaction: {
            from: address,
            to: coinbase,
            value: "123",
            nonce: "10000",
            gas_price: "1000000",
            gas_limit: "1000000",
            type: "deploy",
            contract: {
                "source": "var a = {}",
                "source_type": "ts"
            }
        },
        passphrase: 'passphrase'
    }

    var testExpect = {
        hasError: false,
        errorMsg: "",

    }

    testSignTransaction(testInput, testExpect, done)
});

it('20. `deploy type parse err`', done => {
    var testInput = {
        transaction: {
            from: address,
            to: coinbase,
            value: "123",
            nonce: "10000",
            gas_price: "1000000",
            gas_limit: "1000000",
            type: "deploy",
            contract: {
                "source": "var a = {}"
            }
        },
        passphrase: 'passphrase'
    }

    var testExpect = {
        hasError: true,
        errorMsg: "invalid source type of deploy payload",

    }

    testSignTransaction(testInput, testExpect, done)
});

it('21. `call type`', done => {
    var testInput = {
        transaction: {
            from: address,
            to: coinbase,
            value: "123",
            nonce: "10000",
            gas_price: "1000000",
            gas_limit: "1000000",
            type: "call",
            contract: {
                "function": "save"
            }
        },
        passphrase: 'passphrase'
    }

    var testExpect = {
        hasError: false,
        errorMsg: "",

    }

    testSignTransaction(testInput, testExpect, done)
});
it('22. `call type function err`', done => {
    var testInput = {
        transaction: {
            from: address,
            to: coinbase,
            value: "123",
            nonce: "10000",
            gas_price: "1000000",
            gas_limit: "1000000",
            type: "call",
            contract: {
                "function": "_save"
            }
        },
        passphrase: 'passphrase'
    }

    var testExpect = {
        hasError: true,
        errorMsg: "invalid function of call payload",

    }

    testSignTransaction(testInput, testExpect, done)
});
it('23. `call type no function`', done => {
    var testInput = {
        transaction: {
            from: address,
            to: coinbase,
            value: "123",
            nonce: "10000",
            gas_price: "1000000",
            gas_limit: "1000000",
            type: "call"
        },
        passphrase: 'passphrase'
    }

    var testExpect = {
        hasError: true,
        errorMsg: "invalid function of call payload",

    }

    testSignTransaction(testInput, testExpect, done)
});
});
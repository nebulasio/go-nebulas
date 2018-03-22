'use strict';

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');


var coinbase,
    server_address;

var env = process.env.NET || 'local';
if (env === 'local') {
    server_address = 'localhost:8684';
    coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
} else if (env === 'testneb1') {
    server_address = '35.182.48.19:8684';
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
} else if (env === "testneb2") {
    server_address = "34.205.26.12:8684";
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
}

var client,
    address,
    toAddress;

function testSendTransactionWithPassphrase(testInput, testExpect, done) {
    try {
        client.SendTransactionWithPassphrase(testInput, (err, resp) => {
            try {
                // console.log(JSON.stringify(err));
                expect(!!err).to.equal(testExpect.hasError);

                if (err) {
                    console.log(JSON.stringify(err));
                    expect(err).have.property('details').equal(testExpect.errorMsg);
                } else {
                    console.log(JSON.stringify(resp));
                    expect(resp).to.have.property('txhash');
                    expect(resp).to.have.property('contract_address');
                }
                done();
            } catch (err) {
                done(err);
            }
        });
    } catch (err) {
        try {
            expect(testExpect.hasError).to.be.true;
            expect(err.toString()).to.have.string(testExpect.cannotExecuteError);
            done();
        } catch(er) {
            done(err)
        }
    }
}

describe("rpc: SendTransactionWithPassphrase", () => {
    before((done) => {
        client = rpc_client.new_client(server_address, 'AdminService');
        
        try {
            client.NewAccount({passphrase: "passphrase"}, (err, resp) => {
                expect(!!err).to.be.false;
                expect(resp).to.have.property('address');
                address = resp.address;
                console.log("create new `from` account: " + address);
                
                client.NewAccount({passphrase: "passphraseto"}, (err, resp) => {
                    expect(!!err).to.be.false;
                    expect(resp).to.have.property('address');
                    toAddress = resp.address;
                    console.log("create new `to` account: " + toAddress);
                    done();
                });
            });
        } catch(err) {
            done(err)
        }
    });

    it('1. normal', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: false
        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('2. nonexistent `from`', done => {
        var testInput = {
            transaction: {
                from: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
                to: toAddress,
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
            errorMsg: "address not find",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('3. illegal `from`', done => {
        var testInput = {
            transaction: {
                from: "asfas",
                to: toAddress,
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
            errorMsg: "address: invalid address",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('4. empty `from`', done => {
        var testInput = {
            transaction: {
                from: "",
                to: toAddress,
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
            errorMsg: "address: invalid address",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('5.  illegal `to`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: "afdas",
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
            errorMsg: "address: invalid address",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('6.  empty `to`', done => {
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
            errorMsg: "address: invalid address",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('7. alpha `value', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "23ljfasf",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid value",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('8. negative `value`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "-3242",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid value",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('9. empty `value`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid value",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('10. alpha `nonce`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "abasdx", 
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "transaction's nonce is invalid, should bigger than the from's nonce",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('11. empty `nonce`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "", 
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            cannotExecuteError: 'Error:'
        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('12. negative `nonce`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "-10000", 
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: false,
            cannotExecuteError: ''
        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('13. not bigger than from `nonce`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "0", 
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
         
            errorMsg: "transaction's nonce is invalid, should bigger than the from's nonce",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('14.  alpha `gas_price`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "10000",
                gas_price: "fasf",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasPrice",

        }
        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('15. negative `gas_price`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: 10000,
                gas_price: "-10000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasPrice",

        }
        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('16. empty `gas_price`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "",
                gas_price: "",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "",
            cannotExecuteError: 'Error:'

        }
        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });
    
    it('17.  alpha `gas_limit`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "1233455234",
                nonce: "10000",
                gas_price: "10000",
                gas_limit: "afsdkjkjkkjhf",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasLimit",

        }
        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    // it('16.  alpha `gas_limit` & sufficient `gas_price`', done => {
    //     var testInput = {
    //         transaction: {
    //             from: address,
    //             to: toAddress,
    //             value: "123",
    //             nonce: "10000",
    //             gas_price: "1000000",
    //             gas_limit: "afsdkjkjkkjhf",
    //             contract: {}
    //         },
    //         passphrase: "passphrase"
    //     }

    //     var testExpect = {
    //         hasError: true,

    //     }
    //     testSendTransactionWithPassphrase(testInput, testExpect, done)
    // });

    it('18. negative number `gas_limit`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: 10000,
                gas_price: "10000",
                gas_limit: "-1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasLimit",

        }
        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('19. empty `gas_limit`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "",
                gas_price: "134241",
                gas_limit: "",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "",
            cannotExecuteError: 'Error:'

        }
        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    /* it('20. normal `contract`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,      // contract
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {
                    "function": "save",
                }
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "cannot found account in storage",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('21. empty `contract`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "passphrase"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "duplicated transaction",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    }); */

    it('20. wrong `passphrase`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: "wrongpass"
        }

        var testExpect = {
            hasError: true,
            errorMsg: "could not decrypt key with given passphrase",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });

    it('21. empty `passphrase`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: toAddress,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: ""
        }

        var testExpect = {
            hasError: true,
            errorMsg: "passphrase is invalid",

        }

        testSendTransactionWithPassphrase(testInput, testExpect, done)
    });
});
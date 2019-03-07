"use strict";

var DepositeContent = function (text) {
	if (text) {
		var o = JSON.parse(text);
		this.balance = new BigNumber(o.balance);
		this.expiryHeight = new BigNumber(o.expiryHeight);
	} else {
		this.balance = new BigNumber(0);
		this.expiryHeight = new BigNumber(0);
	}
};

DepositeContent.prototype = {
	toString: function () {
		return JSON.stringify(this);
	}
};

var BankVaultContractS = function () {
	LocalContractStorage.defineMapProperty(this, "bankVault", {
		parse: function (text) {
			return new DepositeContent(text);
		},
		stringify: function (o) {
			return o.toString();
		}
	});
};

// save value to contract, only after height of block, users can takeout
BankVaultContractS.prototype = {
	init: function () {
		//TODO:
	},
	getRandom: function(address, randA) {
		console.log("second bank");
        var rand = _native_math.random();
		console.log("rand_second:", rand);

		var c = new Blockchain.Contract(address);
		var args = "";
		c.value(0).call("getRandom", randA, rand); 
        return rand;
    },
	save: function (address, height) {
		console.log("---------- inner in second address:",address, height);
		var from = Blockchain.transaction.from;
		var value = Blockchain.transaction.value;
		var bk_height = new BigNumber(Blockchain.block.height);

		var orig_deposit = this.bankVault.get(from);
		if (orig_deposit) {
			value = value.plus(orig_deposit.balance);
		}

		var deposit = new DepositeContent();
		deposit.balance = value;
		deposit.expiryHeight = bk_height.plus(height);

		this.bankVault.put(from, deposit);
		this.transferEvent(true, height);
		var c = new Blockchain.Contract(address);
		c.value(2).call("save", height); 
        this.transferEvent(true, height);
	},
	saveMem: function (address, mem) {
        var m = new ArrayBuffer(mem);
		var c = new Blockchain.Contract(address);
        c.value(0).call("saveMem", mem); 
        this.transferEvent(true, 0, mem);
	},
	saveValue: function(address, val) {
        var c = new Blockchain.Contract(address);
        c.value(val).call("saveValue", val);
    },
	saveErr: function(address, flag) {
        if (flag == 1) {
            throw("saveErr in bank_vault_inner_contract");
            return;
        }
        var c = new Blockchain.Contract(address);
        c.value(3).call("saveErr", address, flag); 
        // this.transferEvent(true, address, 0, mem);
	},
	saveTimeOut: function(address, flag) {
		if (flag == 1) {
            while(1) {

			}
        }
        var c = new Blockchain.Contract(address);
        c.value(0).call("saveTimeOut", address, flag); 
	},
	transferEvent: function (status, height, mem) {
        Event.Trigger("bank_vault_contract_second", {
            Status: status,
            BankVaultContractSecond: {
				height: height,
				mem: mem,
                magic: "children one"
            }
        });
    },
	takeout: function (value) {
		var from = Blockchain.transaction.from;
		var bk_height = new BigNumber(Blockchain.block.height);
		var amount = new BigNumber(value);

		var deposit = this.bankVault.get(from);
		if (!deposit) {
			throw new Error("No deposit before.");
		}

		if (bk_height.lt(deposit.expiryHeight)) {
			throw new Error("Can not takeout before expiryHeight.");
		}

		if (amount.gt(deposit.balance)) {
			throw new Error("Insufficient balance.");
		}

		var result = Blockchain.transfer(from, amount);
		if (!result) {
			throw new Error("transfer failed.");
		}
		Event.Trigger("BankVault", {
			Transfer: {
				from: Blockchain.transaction.to,
				to: from,
				value: amount.toString()
			}
		});

		deposit.balance = deposit.balance.sub(amount);
		this.bankVault.put(from, deposit);
	},

	balanceOf: function () {
		var from = Blockchain.transaction.from;
		return this.bankVault.get(from);
	},

	verifyAddress: function (address) {
		// 1-valid, 0-invalid
		var result = Blockchain.verifyAddress(address);
		return {
			valid: result == 0 ? false : true
		};
	}
};

module.exports = BankVaultContractS;

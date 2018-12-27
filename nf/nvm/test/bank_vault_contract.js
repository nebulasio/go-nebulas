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

var BankVaultContract = function () {
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
BankVaultContract.prototype = {
	init: function () {
		//TODO:
	},
	getRandom: function(randA, randB) {
		console.log("second bank");
        var rand = _native_math.random();
		console.log("rand_last:", rand);
		if (rand == randA || rand == randB) {
			throw("check the rand is equal");
		}
		
    },
	save: function (height) {
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
	},
	saveMem: function (mem) {
        var m = new ArrayBuffer(mem);

        this.transferEvent(true, 0, mem);
	},
	saveErr: function(address, flag) {
        if (flag == 2) {
            throw("saveErr in bank_vault_contract");
            return;
        }
        this.transferEvent(true, 0, 3);
	},
	saveValue: function(val) {
		console.log("inner last saveValue:", val);
    },
	saveTimeOut: function(address, flag) {
        if (flag == 2) {
            while(1) {
				
			}
        }
        this.transferEvent(true, 0, 3);
	},
	transferEvent: function (status, height, mem) {
        Event.Trigger("bank_vault_contract", {
            Status: status,
            Transfer: {
				height: height,
				mem: mem,
                magic: "children last one"
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

module.exports = BankVaultContract;

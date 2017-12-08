'use strict';

var DepositeContent = function (text) {
	if (text) {
		let o = JSON.parse(text);
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

	save: function (height) {
		var from = Blockchain.transaction.from;
		var value = Blockchain.transaction.value;
		var bk_height = new BigNumber(Blockchain.block.height);

		var orig_deposit = this.bankVault.get(from);
		if (orig_deposit) {
			value = value.plus(balance);
		}

		var deposit = new DepositeContent();
		deposit.balance = value;
		deposit.expiryHeight = bk_height.plus(height);

		this.bankVault.put(from, deposit);
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
			throw new Error("Can't takeout before expiryHeight.");
		}

		if (amount.gt(deposit.balance)) {
			throw new Error("Insufficient balance.");
		}

		var result = Blockchain.transfer(from, amount);
		if (result != 0) {
			throw new Error("transfer failed.");
		}

		deposit.balance = deposit.balance.sub(amount);
		this.bankVault.put(from, deposit);
	}
};

module.exports = BankVaultContract;

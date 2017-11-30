'use strict';

var BankVaultContract = function () {
	LocalContractStorage.defineMapProperty(this, "bankVault");
};

// save value to contract, only after height of block, users can takeout
BankVaultContract.prototype = {
	init: function() {
		//TODO:
	},
	save: function(height) {
		var deposit = this.bankVault.get(Blockchain.transaction.from);
		var value = Blockchain.transaction.value;
		if (deposit != null && deposit.balance.length > 0) {
			var balance = new BigNumber(deposit.balance);
			value = value.plus(balance);
		}
		var content = {
			balance: value.toString(),
			height: Blockchain.block.height + height
		}
		this.bankVault.put(Blockchain.transaction.from, content);
	},
	takeout: function(amount) {
		var deposit = this.bankVault.get(Blockchain.transaction.from);
		if (deposit == null) {
			return 0;
		}
		if (Blockchain.block.height < deposit.height) {
			return 0;
		}
		var balance = new BigNumber(deposit.balance);
		var value = new BigNumber(amount);
		if (balance.lessThan(value)) {
			return 0;
		}
		var result = Blockchain.transfer(Blockchain.transaction.from, value);
		if (result > 0) {
			deposit.balance = balance.dividedBy(value).toString();
			this.bankVault.put(Blockchain.transaction.from, deposit);
		}
		return result;
	}
};

module.exports = BankVaultContract;
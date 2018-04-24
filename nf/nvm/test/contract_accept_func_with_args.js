
"use strict";

// save value to contract, only after height of block, users can takeout
BankVaultContract.prototype = {
	init: function () {
		//TODO:
	},

	// enables accepting NAS via 'binary'
	accept: function (height) {		
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
	}
};

module.exports = BankVaultContract;


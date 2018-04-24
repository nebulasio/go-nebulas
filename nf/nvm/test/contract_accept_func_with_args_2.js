
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

	// enables accepting NAS via 'binary'
	accept: function (height) {		
		var from = Blockchain.transaction.from;
		var value = Blockchain.transaction.value;
		var bk_height = new BigNumber(Blockchain.block.height);

        console.log("accept height: " + height);

		return {
			height: height,
			from: from,
			value: value,
			bk_height: bk_height
		};
	}
};

module.exports = BankVaultContract;


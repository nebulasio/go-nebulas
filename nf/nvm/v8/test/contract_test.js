// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

// 0. NVM prepare.
const msg = {
    sender: "robin",
    block: {
        number: 102
    }
};
const NebulasToken = require('./contract.js');
var instance = new NebulasToken();

// 1. deploy
instance.init(1000);
console.log("totalSupply = " + token.totalSupply());

// 2. pay.
token.pay(500);
console.log("robin.balance = " + token.balanceOf(msg.sender));
console.log("totalIssued = " + token._totalIssued);

// 3. transfer.
token.transfer("hitters", 200);
console.log("robin.balance = " + token.balanceOf(msg.sender));
console.log("hitters.balance = " + token.balanceOf("hitters"));

// debug.
console.log("dump:");
console.log("token.balances = " + JSON.stringify(token._balances));

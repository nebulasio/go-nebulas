'use strict';

// vote item
var VoteItem = function(dappId, vote) {
    this.dappId = dappId;
    this.vote = vote;
};

var DappVote = function(str) {
    this.total = "0";
    this.voteItems = new Array();
    if (str !== undefined && str !== null) {
        this.parse(str);
    }
};

DappVote.prototype = {
    stringify: function() {
        return JSON.stringify(this);
    },
    parse: function(str) {
        var obj = JSON.parse(str);
        this.total = obj.total;

        var list = obj.voteItems;
        for (var i= 0; i < list.length; i++) {
            var item = list[i];
            var dappVote = new VoteItem(item.dappId, item.vote);
            this.voteItems.push(dappVote);
        }
    },
    vote: function(dappId, vote) {
        if (vote instanceof BigNumber) {
            vote = vote.toString(10);
        }
        var item = new VoteItem(dappId, vote);
        this.voteItems.push(item);
        var totalVote = new BigNumber(this.total);
        this.total = totalVote.add(vote).toString(10);
    },
    voteDappList: function() {
        var list = new Array();
        for(var i = 0; i < this.voteItems.length; i++) {
            list.push(this.voteItems[i].dappId);
        }
        return list;
    }
};

var IncentiveVoteContract = function() {
    LocalContractStorage.defineProperties(this, {
        name: null,     // incentive vote name
        lockAddress: null,  // the lock address of vote
        endHeight: null,   //end of vote height
        owner: null,        // owner
        totalPerVoters: {
            stringify: function(n) {
                return n.toString(10);
            },
            parse: function(str) {
                return new BigNumber(str);
            }
        }
    });

    // the list of dappId
    LocalContractStorage.defineMapProperty(this, "dappArray");
    // the count of dappId list
    LocalContractStorage.defineProperty(this, "dappCount");

    // the list of vote address
    LocalContractStorage.defineMapProperty(this, "addressArray");
    // the count of address list
    LocalContractStorage.defineProperty(this, "addressCount");


    // vote dapp data
    LocalContractStorage.defineMapProperty(this, "voteMap", {
        stringify: function(obj) {
            return JSON.stringify(obj);
        },
        parse: function(str) {
            return new DappVote(str);
        }
    });
    LocalContractStorage.defineMapProperty(this, "voteAddrArray");
    LocalContractStorage.defineProperty(this, "voteCount");
};

IncentiveVoteContract.prototype = {
    init: function(name, lockAddress, endHeight, totalPerVoters, dappArray, addressArray) {
        this.name = name;
        if (Blockchain.verifyAddress(lockAddress) != 0) {
            this.lockAddress = lockAddress;
        } else {
            throw new Error("invalid lock address");
        }
        if (Blockchain.block.height >= endHeight) {
            throw new Error("invalid end height of vote");
        }
        this.endHeight = endHeight;
        this.totalPerVoters = new BigNumber(totalPerVoters);

        for(var i = 0; i < dappArray.length; i++) {
            this.dappArray.set(i, dappArray[i]);
        }
        this.dappCount = dappArray.length;
     
        for(var i = 0; i < addressArray.length; i++) {
            this.addressArray.set(i, addressArray[i]);
        }
        this.addressCount = addressArray.length;

        this.voteCount = 0;

        this.owner = Blockchain.transaction.from;
    },
    getName: function() {
        return this.name;
    },
    vote: function(dappId) {
        if (this.endHeight < Blockchain.block.height) {
            throw new Error("over voting time");
        };
        if (Blockchain.transaction.value.lte(0)) {
            throw new Error("vote must bigger than 0");
        };

        if (!this._isValidDapp(dappId)) {
            throw new Error("invalid vote dapp");
        };
        var from = Blockchain.transaction.from;
        if (!this._isValidVoter(from)) {
            throw new Error("invalid voter address");
        };

        var dappVote = this.voteMap.get(from);
        var isFirstVote = false;
        if (dappVote === null) {
            dappVote = new DappVote();
            isFirstVote = true; 
        };
        var totalVote = new BigNumber(dappVote.total);
        totalVote = totalVote.add(Blockchain.transaction.value);
        if (totalVote.gt(this.totalPerVoters)) {
            throw new Error("out of the count votes per address");
        };

        // vote for the dapp
        dappVote.vote(dappId, Blockchain.transaction.value);
        this.voteMap.put(from, dappVote);
        if (isFirstVote) {
            this.voteAddrArray.set(this.voteCount, from);
            this.voteCount = this.voteCount + 1;
        };
        var result = Blockchain.transfer(this.lockAddress, Blockchain.transaction.value);
        if (!result) {
            throw new Error("vote transfer failed");
        }
    },
    _isValidDapp: function(dappId) {
        var count = this.dappCount;
        for (var i = 0; i < count; i++) {
            if (this.dappArray.get(i) === dappId) {
                return true;
            }
        }
        return false;
    },
    _isValidVoter: function(addr) {
        var count = this.addressCount;
        for (var i = 0; i < count; i++) {
            if (this.addressArray.get(i) === addr) {
                return true;
            }
        }
        return false;
    },
    updateEndHeight: function(height) {
        if (this.owner !== Blockchain.transaction.from) {
            throw new Error("only contract owner can update the end height");
        }
        this.endHeight = height;
    },
    getEndHeight: function() {
        return this.endHeight;
    },
    updateVoter: function(oldAddr, newAddr) {
        if (this.owner !== Blockchain.transaction.from) {
            throw new Error("only contract owner can update the voter");
        }
        if (this.endHeight < Blockchain.block.height) {
            throw new Error("over voting time");
        }
        if (!this._isValidVoter(oldAddr)) {
            throw new Error("invalid old voter address");
        }
        if (this.voteMap.get(oldAddr) !== null) {
            throw new Error("old address has voted dapp, can not be change!");
        }

        if (Blockchain.verifyAddress(newAddr) === 0) {
            throw new Error("invalid lock address");
        };

        var count = this.addressCount;
        for (var i = 0; i < count; i++) {
            if (this.addressArray.get(i) === oldAddr) {
                this.addressArray.set(i, newAddr);
                return;
            }
        }
    },
    getDappList: function() {
        var count = this.dappCount;
        var dappList = new Array();
        for (var i = 0; i < count; i++) {
            dappList.push(this.dappArray.get(i));
        }
        return dappList;
    },
    getVoterList: function() {
        var count = this.addressCount;
        var voterList = new Array();
        for (var i = 0; i < count; i++) {
            voterList.push(this.addressArray.get(i));
        }
        return voterList;
    },
    getAddrVotes: function (addr) {
        if (Blockchain.verifyAddress(addr) === 0) {
            throw new Error("invalid query address");
        };
        if (!this._isValidVoter(addr)) {
            throw new Error("query address is not the voter");
        }
        return this.voteMap.get(addr);
    },
    getAddrVotesList: function() {
        var count = this.addressCount;
        var dappVotes = {};
        for(var i = 0; i < count; i++) {
            var addr = this.addressArray.get(i);
            var dappVote = this.voteMap.get(addr);
            if (dappVote === null) {
                dappVote = new DappVote();
            }
            dappVotes[addr] = dappVote;
        }
        return dappVotes;
    },
    getDappVotes: function() {
        var count = this.voteCount;
        var dappVotes = {};
        for(var i = 0; i < count; i++) {
            var addr = this.voteAddrArray.get(i);
            var dappVote = this.voteMap.get(addr);
            for (var j = 0; j < dappVote.voteItems.length; j++) {
                var voteItem = dappVote.voteItems[j];
                var vote = dappVotes[voteItem.dappId] || "0";
                vote = new BigNumber(vote);
                dappVotes[voteItem.dappId] = vote.add(voteItem.vote).toString(10);
            }
        }

        var count = this.dappCount;
        for (var i = 0; i < count; i++) {
            var dappId = this.dappArray.get(i);
            if (dappVotes[dappId] === undefined) {
                dappVotes[dappId] = "0";
            }
        }
        return dappVotes;
    },
    withdrawal: function(value) {
        if (this.owner !== Blockchain.transaction.from) {
            throw new Error("only contract owner can handle the vote");
        }
        value = new BigNumber(value);
        if (value.lte(0)) {
            throw new Error("value must bigger than 0");
        }
        var result = Blockchain.transfer(this.lockAddress, value);
        if (!result) {
            throw new Error("withdrawal transfer failed");
        }
    },
    accept: function() {
        if (Blockchain.transaction.value.gt(0)) {
            throw new Error("vote contract not accept value without dappId");
        };
    }
};

module.exports = IncentiveVoteContract;
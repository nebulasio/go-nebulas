"use strict";

var WikiItem = function(text) {
  if (text) {
    var obj = JSON.parse(text);
    this.key = obj.key;
    this.value = obj.value;
    this.author = obj.text;
  } else {
      this.key = "";
      this.author = "";
      this.value = "";
  }
};

WikiItem.prototype = {
  toString: function () {
    return JSON.stringify(this);
  }
};

var SuperWiki = function () {
    LocalContractStorage.defineMapProperty(this, "repo", {
        parse: function (text) {
            return new WikiItem(text);
        },
        stringify: function (o) {
            return o.toString();
        }
    });
};

SuperWiki.prototype = {
    init: function () {
        // todo
    },

    save: function (key, value) {
        console.log("reach child contract");
        if(Blockchain.transaction.value < 2000000000000000000) {
            throw("nas is not enough");
        }

        key = key.trim();
        value = value.trim();
        if (key === "" || value === ""){
            throw new Error("empty key / value");
        }
        if (value.length > 128 || key.length > 128){
            throw new Error("key / value exceed limit length")
        }

        var from = Blockchain.transaction.from;
        var wikiItem = this.repo.get(key);
    
        if (wikiItem){
            throw new Error("value has been taken");
        }

        wikiItem = new WikiItem();
        wikiItem.author = from;
        wikiItem.key = key;
        wikiItem.value = value;
        this.repo.put(key, wikiItem);
    },

    get: function (key) {
        key = key.trim();
        if ( key === "" ) {
            throw new Error("empty key")
        }
        return this.repo.get(key);
    }, 

    throwErr: function() {
        throw("err for test");
    }
};
module.exports = SuperWiki;
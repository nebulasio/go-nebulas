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

    testTps: function() {
        console.log("child nvm!");
    },

    testTimeOut: function() {
        while(true){
        }
    },

    testOom: function(){
        var list = new Array();
        console.log("oom-=================");
        while (true) {
            var buffer = new ArrayBuffer(4096);
            var array = new Int32Array(buffer);
            // for (var i = 0; i < 10240; i++) {
            //     array[i] = i;
            // }
            array[1023] = 1;
            list.push(buffer);
        }
    },

    saveWithNoValue: function (key, value) {
        console.log("reach child contract");

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

    save: function (key, value) {
        console.log("reach child contract");
        console.log("==========", key, value);
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

    testTimeOut: function() {
        while(true){
        }
    },

    get: function (key) {
        console.log("child get");
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
'use strict';

var SampleContract = function () {
    LocalContractStorage.defineProperties(this, {
        name: null,
        count: null
    });
    LocalContractStorage.defineMapProperty(this, "allocation");
};

SampleContract.prototype = {
    init: function (mem) {
        var c = new ArrayBuffer(mem);
        console.log("c[1]:", c[1]);
    },
    newMem: function (mem) {
        var c = new ArrayBuffer(mem);
        console.log("c[1]:", c[1]);
    },
    loop: function(){
        var a = 0;
        while(true){
            a +=1;
        };
        return a;
    },
};

module.exports = SampleContract;

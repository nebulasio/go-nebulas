function Set() {
    this.values = {}
    this.length = 0
}

Set.prototype.get = function () {
    return this//支持链式调用
}

Set.prototype.toString = function () {
    ret = this.get() + "test"
    return ret
}

var Test = function () {
};

Test.prototype = {
    init: function () {
	//while(1) {}
       var s = new Set();
       s.toString()
    }
}
var s = new Set();
s.toString();
//while(1) {

//}
console.log("test js");


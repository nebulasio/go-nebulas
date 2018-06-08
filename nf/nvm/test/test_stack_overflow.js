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
       var s = new Set();
       s.toString()
    }
};

module.exports = Test;
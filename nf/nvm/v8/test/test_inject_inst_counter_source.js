'use strict';

var module = {};

var _instruction_counter = 123;

function* countAppleSales() {
    var saleList = [3, 7, 5];
    for (var i = 0; i < saleList.length; i++) {
        yield saleList[i];
    }
}
var appleStore = countAppleSales(); // Generator { }
console.log(appleStore.next()); // { value: 3, done: false }
console.log(appleStore.next()); // { value: 7, done: false }
console.log(appleStore.next()); // { value: 5, done: false }
console.log(appleStore.next()); // { value: undefined, done: true }

var doStatement = function () {
    var i = 0;
    do {
        i++;
    } while (i < 15);

    do
        i++;
    while (i < 25);

    while (i < 100) {
        i++;
        if (i < 50) {
            break;
        }
    }
};

var forStatement = function () {
    for (var i = 0; i < 12; i++) {
        if (i % 3 == 0) {
            continue;
        }
        if (i == 10) {
            break;
        }
    }
};

var arrayFunction = function () {
    let foo = (s) => {
        console.log('a ' + 123);
    };
};

var SampleContract = function () {
    LocalContractStorage.defineProperties(this, {
        name: null,
        count: null
    });
    LocalContractStorage.defineMapProperty(this, "allocation" + 123);
    this.a = 0;
    let foo = (s) => s + 1;
    new.target;
    var elvisLives = Math.PI > 4 ? foo() : 'Nope';
};
SampleContract.prototype = {
    init: function (name, count, allocation) {
        this.name = name;
        this.count = count;
        this.zz[0] = 123;
        allocation.forEach(function (item) {
            this.allocation.put(item.name, item.count);
        }, this);
    },
    dump: function () {
        console.log('dump: this.name = ' + this.name);
        console.log('dump: this.count = ' + this.count);
        return this.a;
    },
    incr: function () {
        this.a++;
        var z = this.a;
        console.log(this.dump());
        return this.a;
    },
    verify: function (expectedName, expectedCount, expectedAllocation) {
        if (!Object.is(this.name, expectedName))
            throw new Error("name is not the same, expecting " + expectedName + ", actual is " + this.name + ".");
        else
            var elvisLives = Math.PI > 4 ? foo() : 'Nope';

        if (!Object.is(this.count, expectedCount)) {
            throw new Error("count is not the same, expecting " + expectedCount + ", actual is " + this.count + ".");
        } else {
            console.log('ok.');
        }
        expectedAllocation.forEach(function (expectedItem) {
            var count = this.allocation.get(expectedItem.name);
            if (!Object.is(count, expectedItem.count)) {
                throw new Error("count of " + expectedItem.name + " is not the same, expecting " + expectedItem.count + ", actual is " + count + ".");
            }
        }, this);
    },
    test_switch: function () {
        var day = "";
        switch (new Date().getDay()) {
            case 0:
                day = "Sunday";
                break;
            case 1:
                day = "Monday";
                break;
            case 2:
                day = "Tuesday";
                break;
            case 3:
                day = "Wednesday";
                break;
            case 4:
                day = "Thursday";
                break;
            case 5:
                day = "Friday";
                break;
            case 6:
                day = "Saturday";
            default:
                throw new Error("N/A");
        }
        console.log('day is ' + day);
    },
    test_for: function () {
        for (var i = 0; i < 123; i++) {
            var z = i;
            alert(z);
        }
        for (var i = 0; i < 123; i++)
            alert(i);
    },
};

module.exports = SampleContract;

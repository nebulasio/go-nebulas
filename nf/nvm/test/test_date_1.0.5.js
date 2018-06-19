// Copyright (C) 2018 go-nebulas authors
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

function eq(a, b) {
    if (a !== b) {
        throw new Error("Not equal: " + a + " <--> " + b);
    }
}

Blockchain.blockParse("{\"timestamp\":20000000000,\"seed\":\"\"}");

var date = new Date();
var options = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' };

eq(date.toDateString(), "Tue Oct 11 2603");
eq(Date.UTC(2603, 9, 11, 11, 33, 20), 20000000000000);
eq(date.getTimezoneOffset(), 0);
eq(date.toTimeString(), "11:33:20 GMT+0000 (UTC)");
eq(date.toString(), "Tue Oct 11 2603 11:33:20 GMT+0000 (UTC)");
eq(date.toGMTString(), "Tue, 11 Oct 2603 11:33:20 GMT");
eq(date.toUTCString(), "Tue, 11 Oct 2603 11:33:20 GMT");
eq(date.toISOString(), "2603-10-11T11:33:20.000Z");
eq(date.toJSON(), "2603-10-11T11:33:20.000Z");
eq(date.valueOf(), 20000000000000);

eq(Object.prototype.toLocaleString.call(date), "Tue Oct 11 2603 11:33:20 GMT+0000 (UTC)");
eq(Object.prototype.toLocaleString.call(date, 'ko-KR', { timeZone: 'UTC' }), "Tue Oct 11 2603 11:33:20 GMT+0000 (UTC)");
eq(date.toLocaleString(), "10/11/2603, 11:33:20 AM");
eq(date.toLocaleString('ko-KR', { timeZone: 'UTC' }), "2603. 10. 11. 오전 11:33:20");

eq(date.toLocaleDateString(), "10/11/2603");
eq(date.toLocaleDateString('de-DE', options), "Dienstag, 11. Oktober 2603");

eq(date.toLocaleTimeString(), "11:33:20 AM");
eq(date.toLocaleTimeString('ar-EG'), "١١:٣٣:٢٠ ص");


eq(date.getDate(), date.getUTCDate());
eq(date.getDay(), date.getUTCDay());
eq(date.getFullYear(), date.getUTCFullYear());
eq(date.getHours(), date.getUTCHours());
eq(date.getMilliseconds(), date.getUTCMilliseconds());
eq(date.getMinutes(), date.getUTCMinutes());
eq(date.getMonth(), date.getUTCMonth());
eq(date.getSeconds(), date.getUTCSeconds());

try {
    date.getYear();
    throw new Error("should not be here.");
} catch(err) {
    if (err != "Error: Deprecated!") {
        throw err;
    }
}
try {
    date.setYear(1999);
    throw new Error("should not be here.");
} catch(err) {
    if (err != "Error: Deprecated!") {
        throw err;
    }
}

date = new Date('August 19, 2017 23:15:30');
var date2 = new Date('August 19, 2017 23:15:30');

date.setDate(12);
date2.setUTCDate(12);
eq(date - date2, 0);

date.setMonth(1);
date2.setUTCMonth(1);
eq(date - date2, 0);

date.setFullYear(1999);
date2.setUTCFullYear(1999);
eq(date - date2, 0);

date.setHours(22);
date2.setUTCHours(22);
eq(date - date2, 0);

date.setMilliseconds(420);
date2.setUTCMilliseconds(420);
eq(date - date2, 0);

date.setMinutes(12);
date2.setUTCMinutes(12);
eq(date - date2, 0);

date.setSeconds(12);
date2.setUTCSeconds(12);
eq(date - date2, 0);



var d1 = new Date('December 17, 1995 00:00:00');
var d2 = new Date(d1.getUTCFullYear(), d1.getUTCMonth(), d1.getUTCDate());
eq(d1.getTime(), d2.getTime());

eq(Date.parse(Date()), 20000000000000);
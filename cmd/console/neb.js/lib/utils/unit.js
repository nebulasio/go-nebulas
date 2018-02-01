
"use strict";

var BigNumber = require('bignumber.js');
var utils = require('./utils.js');

var unitMap = {
    'none':       '0',
    'nis':        '1',
    'knis':       '1000',
    'mnis':       '1000000',
    'nanonas':    '1000000000',
    'micronas':   '1000000000000',
    'millinas':   '1000000000000000',
    'nas':        '1000000000000000000',
 };

var unitValue = function (unit) {
	unit = unit ? unit.toLowerCase() : 'nas';
    var unitValue = unitMap[unit];
    if (unitValue === undefined) {
        throw new Error('The unit undefined, please use the following units:' + JSON.stringify(unitMap, null, 2));
    }
    return new BigNumber(unitValue, 10);
};

var toBasic = function (number, unit) {
	return utils.toBigNumber(number).times(unitValue(unit));
};

var fromBasic = function (number, unit) {
	return utils.toBigNumber(number).dividedBy(unitValue(unit));
};

var nasToBasic = function (number) {
	return utils.toBigNumber(number).times(unitValue("nas"));
};

module.exports = {
	toBasic: toBasic,
	fromBasic: fromBasic,
	nasToBasic: nasToBasic
};

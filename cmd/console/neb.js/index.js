var Neb = require('./lib/neb');

// dont override global variable
if (typeof window !== 'undefined' && typeof window.Neb === 'undefined') {
    window.Neb = Neb;
}

module.exports = Neb;

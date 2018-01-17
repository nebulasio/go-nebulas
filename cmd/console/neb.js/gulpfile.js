#!/usr/bin/env node

'use strict';

var version = require('./lib/version.json');
var path = require('path');

var del = require('del');
var gulp = require('gulp');
var browserify = require('browserify');
var jshint = require('gulp-jshint');
var uglify = require('gulp-uglify');
var rename = require('gulp-rename');
var plumber = require('gulp-plumber');
var notify  = require('gulp-notify');
var source = require('vinyl-source-stream');
var exorcist = require('exorcist');
var streamify = require('gulp-streamify');
var replace = require('gulp-replace');
var babelify     = require('babelify');
var html2js      = require('html2js-browserify');
var buffer       = require('vinyl-buffer');

var DEST = path.join(__dirname, 'dist/');
var src = 'index';
var dst = 'neb';
var lightDst = 'neb-light';
var accountDst = 'account';
var transactionDst = 'transaction';
var walletDst = 'wallet';

// Error / Success Handling
var onError = function(err) {
    notify.onError({
        title: "Error: " + err.plugin,
        subtitle: "<%= file.relative %>",
        message: "<%= error.message %>",
        sound: "Beep",
    })(err);
    console.log(err.toString())
    this.emit('end');
}

function onSuccess(msg) {
    return {
        message: msg + " Complete! ",
        onLast: true
    }
}

function notifyFunc(msg) {
    return gulp.src('.', { read: false })
        .pipe(notify(onSuccess(msg)))
}

var browserifyOptions = {
    debug: true,
    insert_global_vars: false, // jshint ignore:line
    detectGlobals: false,
    bundleExternal: true
};

gulp.task('version', function(){
  gulp.src(['./package.json'])
    .pipe(replace(/\"version\"\: \"([\.0-9]*)\"/, '"version": "'+ version.version + '"'))
    .pipe(gulp.dest('./'));
});

gulp.task('lint', [], function(){
    return gulp.src(['./index.js', './lib/*.js', './lib/**/*.js'])
        .pipe(jshint())
        .pipe(jshint.reporter('default'));
});

gulp.task('clean', ['lint'], function(cb) {
    del([ DEST ]).then(cb.bind(null, null));
});

gulp.task('light', ['clean'], function () {
    return browserify(browserifyOptions)
        .require('./' + src + '.js', {expose: 'neb'})
        .ignore('bignumber.js')
        .require('./lib/utils/browser-bignumber.js', {expose: 'bignumber.js'}) // fake bignumber.js
        .add('./' + src + '.js')
        .bundle()
        .pipe(exorcist(path.join( DEST, lightDst + '.js.map')))
        .pipe(source(lightDst + '.js'))
        .pipe(gulp.dest( DEST ))
        .pipe(streamify(uglify()))
        .pipe(rename(lightDst + '.min.js'))
        .pipe(gulp.dest( DEST ));
});

gulp.task('neb', ['clean'], function () {
    return browserify(browserifyOptions)
        .require('./' + src + '.js', {expose: 'neb'})
        .add('./' + src + '.js')
        .bundle()
        .pipe(exorcist(path.join( DEST, dst + '.js.map')))
        .pipe(source(dst + '.js'))
        .pipe(gulp.dest( DEST ))
        .pipe(streamify(uglify()))
        .pipe(rename(dst + '.min.js'))
        .pipe(gulp.dest( DEST ));
});

gulp.task('account', ['clean'], function () {
    return browserify()
        .require('./lib/account.js', {expose: 'account'})
        .transform(babelify)
        .transform(html2js)
        .bundle()
        .pipe(plumber({ errorHandler: onError }))
        // .pipe(exorcist(path.join( DEST, accountDst + '.js.map')))
        .pipe(source('account.js'))
        .pipe(buffer())
        .pipe(rename(accountDst + '.js'))
        .pipe(gulp.dest(DEST));
});

gulp.task('transaction', ['clean'], function () {
    return browserify()
        .require('./lib/transaction.js', {expose: 'transaction'})
        .transform(babelify)
        .transform(html2js)
        .bundle()
        .pipe(plumber({ errorHandler: onError }))
        // .pipe(exorcist(path.join( DEST, accountDst + '.js.map')))
        .pipe(source('transaction.js'))
        .pipe(buffer())
        .pipe(rename(transactionDst + '.js'))
        .pipe(gulp.dest(DEST));
});

gulp.task('wallet', ['clean'], function () {
    return browserify()
        .require('./lib/wallet.js', {expose: 'wallet'})
        .transform(babelify)
        .transform(html2js)
        .bundle()
        .pipe(plumber({ errorHandler: onError }))
        // .pipe(exorcist(path.join( DEST, accountDst + '.js.map')))
        .pipe(source('wallet.js'))
        .pipe(buffer())
        .pipe(rename(walletDst + '.js'))
        .pipe(gulp.dest(DEST));
});

gulp.task('watch', function() {
    gulp.watch(['./lib/*.js'], ['lint', 'build']);
});

gulp.task('default', ['version', 'lint', 'clean', 'light', 'neb', 'account', 'transaction', 'wallet']);


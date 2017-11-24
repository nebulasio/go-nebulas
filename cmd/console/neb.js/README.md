# Nebulas JavaScript API


This is the Nebulas compatible JavaScript API.Users can use it in brower and node.js.This javascript library also support API for our Repl console. 

## Build
We need to use NPM to install dependencies before using:

```
cd <path>/neb.js
npm install
```
We build neb.js/neb-light.js by [gulp](https://gulpjs.com/):

```
gulp
```
Build neb-node.js by [rollup](https://rollupjs.org/):

```
rollup -c
```
Our ouput library in `/dist/` document.

## Library

 * `neb.js`:Used in brower side. Including outside dependency.
 * `neb-light.js`: Used in brower side and Repl console. Ignore outside dependency.
 * `neb-node.js`: Used in node.js. Including outside dependency.
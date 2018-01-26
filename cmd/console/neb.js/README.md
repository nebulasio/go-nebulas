# Nebulas JavaScript API

This readme will get you started with using the `neb.js` file.
This file can be found in go-nebulas/cmd/console/neb.js/dist/neb.js

This is the Nebulas compatible JavaScript API. Users can use it in brower and node.js.This javascript library also support API for our Repl console. 


## Step 1: We need to use NPM to install the dependencies before using neb.js:

Download Node.js:

Note: npm is a Node.js package manager. So this means you need to install Node.js in order to use `npm install` [Node.js](https://nodejs.org/en/)

Once you have Node.js installed we can go to the neb.js folder.

Example 1: in the terminal

```
cd go-nebulas/cmd/console/neb.js
npm install
```


## Step 2: We need to use the `gulp` command

Download Gulp:

Note: In order to use the `gulp` command we need to install [gulp](https://gulpjs.com/):

Example 2 in the terminal:

```
cd go-nebulas/cmd/console/neb.js
gulp
```

Now we can check the newly created files in go-nebulas/cmd/console/neb.js/dist

```
cd go-nebulas/cmd/console/neb.js/dist
```

![image](https://user-images.githubusercontent.com/21117852/35438944-a7e5e2ca-02d3-11e8-84fe-f0987b4b44b8.png)

Here you should see a bunch of js files. 

In the next tutorial I will show how to use neb.js file and make a make a program to interact with your local Nebulas installation.


## Extra 
Rollup is a module bundler for JavaScript which compiles small pieces of code into something larger and more complex, such as a library or application.

Build neb-node.js by [rollup](https://rollupjs.org/):

```
rollup -c
```


## Library

 * `neb.js`:Used in brower side. Including outside dependency.
 * `neb-light.js`: Used in brower side and Repl console. Ignore outside dependency.
 * `neb-node.js`: Used in node.js. Including outside dependency.

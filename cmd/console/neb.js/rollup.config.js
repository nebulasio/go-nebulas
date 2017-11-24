import commonjs from 'rollup-plugin-commonjs';
import nodeResolve from 'rollup-plugin-node-resolve';

export default {
  entry: './index.js',
  dest: 'dist/neb-node.js',
  format: 'cjs',
  plugins: [

    commonjs({
        // non-CommonJS modules will be ignored, but you can also
        // specifically include/exclude files
        include: ['./index.js', './lib/**', "node_modules/**"],  // Default: undefined

        // if true then uses of `global` won't be dealt with by this plugin
        ignoreGlobal: false,  // Default: false

        // if false then skip sourceMap generation for CommonJS modules
        sourceMap: false,  // Default: true
    }),

    nodeResolve({
    jsnext: true,
    main: false
    })
]
};
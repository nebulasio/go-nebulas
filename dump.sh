#!/bin/bash

if [[ "$TRAVIS_OS_NAME" == "linux" ]]; then
/usr/bin/ldd /usr/local/lib/libv8engine.so
ls -alh /usr/lib/x86_64-linux-gnu/libstdc++.*
strings /usr/lib/x86_64-linux-gnu/libstdc++.so.6 | grep GCC
strings /usr/lib/x86_64-linux-gnu/libstdc++.so.6 | grep GLIBCXX
fi

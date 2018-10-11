
#CUR_DIR="$( pwd )"
CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
export AR=$CUR_DIR/lib/bin/llvm-ar
export CXX=$CUR_DIR/lib/bin/clang++
export CC=$CUR_DIR/lib/bin/clang
export PATH=$CUR_DIR/lib/bin:$PATH

case "$(uname -s)" in
'Linux')
    export LD_LIBRARY_PATH=$CUR_DIR/lib/lib:$LD_LIBRARY_PATH
    ;;
'Darwin')
    if [ ! -d ~/lib ]; then
        ln -s $CUR_DIR/lib/lib ~/lib
    fi
    ;;
*) ;;
esac

export NBRE_DB=$CUR_DIR/test/data/write-data.db
export BLOCKCHAIN_DB=$CUR_DIR/test/data/read-data.db

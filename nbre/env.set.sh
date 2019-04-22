
#CUR_DIR="$( pwd )"
CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
export AR=$CUR_DIR/lib_llvm/bin/llvm-ar
export CXX=$CUR_DIR/lib_llvm/bin/clang++
export CC=$CUR_DIR/lib_llvm/bin/clang
export PATH=$CUR_DIR/lib/bin:$CUR_DIR/lib_llvm/bin:$PATH

case "$(uname -s)" in
'Linux')
    export LD_LIBRARY_PATH=$CUR_DIR/lib/lib:$CUR_DIR/lib_llvm/lib:$CUR_DIR/bin:$LD_LIBRARY_PATH
    ;;
'Darwin')
    if [ ! -d ~/lib ]; then
        ln -s $CUR_DIR/lib/lib ~/lib
    fi
    ;;
*) ;;
esac

USER=`whoami`
NEBULAS_HOME=$CUR_DIR/../

export NBRE_ROOT_DIR=$NEBULAS_HOME/nbre
export NBRE_EXE_NAME=$NBRE_ROOT_DIR/nbre
export NEB_DB_DIR=$NEBULAS_HOME/data.db
export NBRE_DB_DIR=$NBRE_ROOT_DIR/nbre.db
export NBRE_LOG_DIR=$NBRE_ROOT_DIR/logs

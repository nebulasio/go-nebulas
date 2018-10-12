#!/bin/bash

# Copyright (C) 2018 go-nebulas authors
#
# This file is part of the go-nebulas library.
#
# the go-nebulas library is free software: you can redistribute it and/or
# modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# the go-nebulas library is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with the go-nebulas library.  If not, see
# <http://www.gnu.org/licenses/>.
#

USE_CASE_ARRAY=("nasir" "blockgen" "naxer")
PROGRAM_SELF_NAME="valgrind_test_nbre.sh"
VALGRIND_ARGUMENTS="--leak-check=yes"
VALGRIND_OUTPUT="./valgrind_report"

show_test_cases()
{
 echo "Test Case Name:"
 for i in "${USE_CASE_ARRAY[@]}"
 do
   echo -e "\t ${i} "
 done
}

usage()
{
  echo "Usage:"
  echo -e "\t$PROGRAM_SELF_NAME -h"
  echo -e "\t$PROGRAM_SELF_NAME -l"
  echo -e "\t$PROGRAM_SELF_NAME -t <test case name>"
  echo -e "\t$PROGRAM_SELF_NAME -o <output_report_path> -t <test case name>"
  echo -e "\t$PROGRAM_SELF_NAME -a"
  echo -e "Parameter:"
  echo -e "\t-h: help"
  echo -e "\t-l: list name of test case"
  echo -e "\t-t: test"
  echo -e "\t-o: output path"
  echo -e "\t-a: run all test case"
  exit 1
}

check_environment()
{
  if [ ! -x "/usr/bin/valgrind" ]; then
    echo "Valgrind don't exist, please execute: sudo apt install valgrind"
    exit 1
  fi

  if [ ! -x "./"$1 ]; then
    echo "Program $1 don't exist!"
    exit 1
  fi
}

check_test_cases() {
  for i in "${USE_CASE_ARRAY[@]}"
  do
    if [ $1 = $i ]; then
      return 0
    fi
  done

  echo "Test case don't exist!"
  exit 1
}

execute_test()
{
  check_environment
  check_test_cases $1

  mkdir -p $VALGRIND_OUTPUT

  if [ $1 = "nasir" ]; then
    echo "Valgrind is checking memory leak of "$1
    valgrind $VALGRIND_ARGUMENTS ./nasir \
      --input ../test/data/test_nasir.json \
      --output nr.bc > $VALGRIND_OUTPUT"/"$1_report 2>&1
  elif [ $1 = "blockgen" ]; then
    echo "Valgrind is checking memory leak of "$1
    valgrind $VALGRIND_ARGUMENTS ./blockgen \
      --ir_binary ./nr.bc \
      --block_conf ../test/data/test_blockgen.json > $VALGRIND_OUTPUT"/"$1_report 2>&1
  elif [ $1 = "naxer" ]; then
    echo "Valgrind is checking memory leak of "$1
    valgrind $VALGRIND_ARGUMENTS ./naxer \
      --module nr \
      --height 1000 > $VALGRIND_OUTPUT"/"$1_report 2>&1
  else
    echo "nothing"
  fi
}

execute_all_test()
{
  for i in "${USE_CASE_ARRAY[@]}"
  do
    execute_test $i
  done
}

if [ $# = 0 ]; then
  usage
  exit 1
fi

while getopts "l t:o: a h" opt; do
  case "$opt" in
    h)
      usage
      ;;
    l)
      show_test_cases
      ;;
    t)
      execute_test $OPTARG
      ;;
    o)
      VALGRIND_OUTPUT=$OPTARG
      ;;
    a)
      execute_all_test
      ;;
    *)
      echo "ERROR: unknow parameter \"$opt\""
      usage
      ;;
  esac
done
shift $((OPTIND-1))




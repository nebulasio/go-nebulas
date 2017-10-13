// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//

#include "engine.h"
#include <stdio.h>
#include <stdlib.h>

void LogError(const char *msg) { printf("Error: %s\n", msg); }

extern "C" int roll_dice() { return 123; }

void help(const char *name) {
  printf("%s [IR file] [Func Name] [Args...]\n", name);
  exit(1);
}

int main(int argc, const char *argv[]) {
  if (argc < 3) {
    help(argv[0]);
  }

  const char *filePath = argv[1];
  const char *funcName = argv[2];

  Initialize();
  printf("initialized.\n");

  Engine *e = CreateEngine(filePath);
  printf("engine created.\n");

  int ret = RunFunction(e, funcName, 0, NULL);
  printf("runFunction return %d\n", ret);

  DeleteEngine(e);
  printf("engine deleted.\n");

  return ret;
}

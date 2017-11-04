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
#include "lib/log_callback.h"
#include "lib/memory_storage.h"

#include <stdio.h>
#include <stdlib.h>

static void *lcsHandler = NULL;
static void *gcsHandler = NULL;

void logFunc(int level, const char *msg) {
  FILE *f = stdout;
  if (level >= LogLevel::ERROR) {
    f = stderr;
  }
  fprintf(f, "[%s] %s\n", GetLogLevelText(level), msg);
}

void help(const char *name) {
  printf("%s [Javascript File] [Args...]\n", name);
  exit(1);
}

void readSource(const char *filename, char **data, size_t *size) {
  FILE *f = fopen(filename, "r");
  if (f == NULL) {
    fprintf(stderr, "file %s does not found.\n", filename);
    exit(1);
    return;
  }

  *size = 1024;
  *data = (char *)malloc(*size);

  size_t len = 0;
  size_t idx = 0;
  while ((len = fread(*data + idx, sizeof(char), *size - idx, f)) > 0) {
    idx += len;
    if (*size - idx <= 1) {
      *size *= 1.5;
      *data = (char *)realloc(*data, *size);
    }
  }
  *(*data + idx) = '\0';

  if (feof(f) == 0) {
    fprintf(stderr, "read file %s error.\n", filename);
    exit(1);
  }

  fclose(f);
}

int main(int argc, const char *argv[]) {
  if (argc < 2) {
    help(argv[0]);
  }

  const char *filename = argv[1];
  char *data = NULL;
  size_t size = 0;
  readSource(filename, &data, &size);

  // temp set handler pointer.
  lcsHandler = (void *)filename;
  gcsHandler = (void *)data;

  Initialize();
  InitializeLogger(logFunc);
  InitializeStorage(StorageGet, StoragePut, StorageDel);

  V8Engine *e = CreateEngine();
  int ret = RunScriptSource(e, data, lcsHandler, gcsHandler);
  DeleteEngine(e);

  Dispose();
  free(data);

  return ret;
}

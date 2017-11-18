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
#include "lib/blockchain.h"
#include "lib/fake_blockchain.h"
#include "lib/log_callback.h"
#include "lib/logger.h"
#include "lib/memory_storage.h"

#include <thread>
#include <vector>

#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static int concurrency = 1;
static int enable_tracer_injection = 0;
static size_t limits_of_executed_instructions = 0;
static size_t limits_of_total_memory_size = 0;
static int print_injection_result = 0;

void logFunc(int level, const char *msg) {
  std::thread::id tid = std::this_thread::get_id();
  std::hash<std::thread::id> hasher;

  FILE *f = stdout;
  if (level >= LogLevel::ERROR) {
    f = stderr;
  }
  fprintf(f, "[tid-%020zu] [%s] %s\n", hasher(tid), GetLogLevelText(level),
          msg);
}

void help(const char *name) {
  printf("%s [-c <concurrency>] [-i] [-li <number>] [-lm <number>] <Javascript "
         "File>\n",
         name);
  printf("\t -c <concurrency> \tNumber of multiple thread to run at a time.\n");
  printf("\t -i \tinject tracing code.\n");
  printf("\t -li <number> \tlimits of executed instructions, default is 0 "
         "(unlimited).\n");
  printf("\t -im <number> \tlimits of total heap size, default is 0 "
         "(unlimited).\n");
  printf("\n");

  printf("%s -ip <Javascript File>\n", name);
  printf("\t -ip \tinject tracing code and print result\n");
  printf("\n");
  exit(1);
}

void readSource(const char *filename, char **data, size_t *size) {
  FILE *f = fopen(filename, "r");
  if (f == NULL) {
    fprintf(stderr, "file %s does not found.\n", filename);
    exit(1);
    return;
  }

  // get file size.
  fseek(f, 0L, SEEK_END);
  size_t fileSize = ftell(f);
  rewind(f);

  *size = fileSize + 1;
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

typedef void (*V8ExecutionDelegate)(V8Engine *e, const char *data,
                                    uintptr_t lcsHandler, uintptr_t gcsHandler);

void RunScriptSourceDelegate(V8Engine *e, const char *data,
                             uintptr_t lcsHandler, uintptr_t gcsHandler) {
  if (enable_tracer_injection) {
    e->limits_of_executed_instructions = limits_of_executed_instructions;
    e->limits_of_total_memory_size = limits_of_total_memory_size;

    char *traceableSource = InjectTracingInstructions(e, data);
    if (traceableSource == NULL) {
      fprintf(stderr, "Inject tracing instructions failed.\n");
    } else {
      int ret = RunScriptSource(e, traceableSource, (uintptr_t)lcsHandler,
                                (uintptr_t)gcsHandler);
      free(traceableSource);

      fprintf(stdout, "[V8] Execution ret = %d\n", ret);

      ret = IsEngineLimitsExceeded(e);
      if (ret) {
        fprintf(stdout, "[V8Error] Exceed %s limits, ret = %d\n",
                ret == 1 ? "Instructions" : "Memory", ret);
      }

      // print tracing stats.
      fprintf(stdout,
              "\nStats of V8Engine:\n"
              "  count_of_executed_instructions: \t%zu\n"
              "  total_memory_size: \t\t\t%zu\n"
              "  total_heap_size: \t\t\t%zu\n"
              "  total_heap_size_executable: \t\t%zu\n"
              "  total_physical_size: \t\t\t%zu\n"
              "  total_available_size: \t\t%zu\n"
              "  used_heap_size: \t\t\t%zu\n"
              "  heap_size_limit: \t\t\t%zu\n"
              "  malloced_memory: \t\t\t%zu\n"
              "  peak_malloced_memory: \t\t%zu\n"
              "  total_array_buffer_size: \t\t%zu\n"
              "  peak_array_buffer_size: \t\t%zu\n",
              e->stats.count_of_executed_instructions,
              e->stats.total_memory_size, e->stats.total_heap_size,
              e->stats.total_heap_size_executable, e->stats.total_physical_size,
              e->stats.total_available_size, e->stats.used_heap_size,
              e->stats.heap_size_limit, e->stats.malloced_memory,
              e->stats.peak_malloced_memory, e->stats.total_array_buffer_size,
              e->stats.peak_array_buffer_size);
    }
  } else {
    RunScriptSource(e, data, (uintptr_t)lcsHandler, (uintptr_t)gcsHandler);
  }
}

void InjectTracingInstructionsAndPrintDelegate(V8Engine *e, const char *data,
                                               uintptr_t lcsHandler,
                                               uintptr_t gcsHandler) {
  char *traceableSource = InjectTracingInstructions(e, data);
  if (traceableSource == NULL) {
    fprintf(stderr, "Inject tracing instructions failed.\n");
  } else {
    fprintf(stdout, "%s", traceableSource);
    free(traceableSource);
  }
}

void ExecuteScript(const char *data, V8ExecutionDelegate delegate) {
  void *lcsHandler = CreateStorageHandler();
  void *gcsHandler = CreateStorageHandler();

  V8Engine *e = CreateEngine();
  delegate(e, data, (uintptr_t)lcsHandler, (uintptr_t)gcsHandler);
  DeleteEngine(e);

  DeleteStorageHandler(lcsHandler);
  DeleteStorageHandler(gcsHandler);
}

void ExecuteScriptSource(const char *data) {
  ExecuteScript(data, RunScriptSourceDelegate);
}

int main(int argc, const char *argv[]) {
  if (argc < 2) {
    help(argv[0]);
  }

  Initialize();
  InitializeLogger(logFunc);
  InitializeStorage(StorageGet, StoragePut, StorageDel);
  InitializeBlockchain(GetBlockByHash, GetTxByHash, GetAccountState, Transfer,
                       VerifyAddress);

  int argcIdx = 1;
  const char *filename = NULL;

  for (;;) {
    const char *arg = argv[argcIdx];
    if (strcmp(arg, "-c") == 0) {
      argcIdx++;
      concurrency = atoi(argv[argcIdx]);
      argcIdx++;
      if (concurrency <= 0) {
        fprintf(stderr, "concurrency can't less than 0, set to 1.\n");
        concurrency = 1;
      }
    } else if (strcmp(arg, "-i") == 0) {
      argcIdx++;
      enable_tracer_injection = 1;
    } else if (strcmp(arg, "-li") == 0) {
      argcIdx++;

      char *s = NULL;
      long limits = strtol(argv[argcIdx], &s, 10);
      argcIdx++;

      if (errno == EINVAL) {
        continue;
      }

      if (errno == ERANGE) {
        // do nothing.
        limits_of_executed_instructions = 0;
      } else {
        limits_of_executed_instructions = limits;
      }
    } else if (strcmp(arg, "-lm") == 0) {
      argcIdx++;

      char *s = NULL;
      long limits = strtol(argv[argcIdx], &s, 10);
      argcIdx++;

      if (errno == EINVAL) {
        continue;
      }

      if (errno == ERANGE) {
        // do nothing.
        limits_of_total_memory_size = 0;
      } else {
        limits_of_total_memory_size = limits;
      }
    } else if (strcmp(arg, "-ip") == 0) {
      argcIdx++;
      print_injection_result = 1;
    } else {
      filename = arg;
      break;
    }
  }

  char *data = NULL;
  size_t size = 0;
  readSource(filename, &data, &size);

  if (print_injection_result) {
    // inject and print.
    ExecuteScript(data, InjectTracingInstructionsAndPrintDelegate);
  } else {
    // execute script.
    std::vector<std::thread *> threads;
    for (int i = 0; i < concurrency; i++) {
      std::thread *thread = new std::thread(ExecuteScriptSource, data);
      threads.push_back(thread);
    }

    for (int i = 0; i < concurrency; i++) {
      threads[i]->join();
    }
  }

  free(data);

  Dispose();
  return 0;
}

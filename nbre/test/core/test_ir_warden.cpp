#include "core/ir_warden.h"
#include "fs/util.h"

int main(int argc, char *argv[]) {

  auto &instance = neb::core::ir_warden::instance();
  instance.release();
  return 0;
}

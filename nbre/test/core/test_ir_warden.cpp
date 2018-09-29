#include "core/ir_warden.h"
#include "fs/util.h"

int main(int argc, char *argv[]) {
  neb::util::singleton_guard<neb::core::ir_warden> _l;
  neb::core::ir_warden::instance();
  return 0;
}

#include "core/ir_warden.h"
#include <gtest/gtest.h>

TEST(test_core_simple, simple) { neb::core::ir_warden::instance().async_run(); }

/**
 * Copyright (C) 2017 go-nebulas authors
 *
 * This file is part of the go-nebulas library.
 *
 * the go-nebulas library is free software: you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * the go-nebulas library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with the go-nebulas library.  If not, see
 * <http://www.gnu.org/licenses/>.
 *
 */

// This is a sample smart contract written in C.

#include <stdio.h>

int roll_dice();

void func_a() { printf("called to func_a.\n"); }

void func_b() {
  int v = roll_dice();
  printf("called to func_b, dice is %d.\n", v);
}

int main(int argc, char *argv[]) {
  printf("called to main.\n");
  printf("argc = %d\n", argc);
  func_a();
  func_b();
  return 234;
}

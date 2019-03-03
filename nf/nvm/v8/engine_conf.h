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
#pragma once

// Basic configuration and settings of the V8 engine
const int ExecutionFailedErr  = 1;
const int ExecutionTimeOutErr = 2;

typedef unsigned long long uint64;

// ExecutionTimeoutInSeconds max v8 execution timeout.
const uint64 ExecutionTimeout                 = 15 * 1000 * 1000;
const uint64 TimeoutGasLimitCost              = 100000000;


// DefaultLimitsOfTotalMemorySize default limits of total memory size
const uint64 MaxLimitsOfExecutionInstructions = 10000000; // 10,000,000
const uint64 DefaultLimitsOfTotalMemorySize = 40 * 1000 * 1000;
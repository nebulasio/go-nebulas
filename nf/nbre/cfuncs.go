// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package nbre

/*

#include <stdint.h>

void NbreVersionFunc(int isc, void *holder, uint32_t major, uint32_t minor,uint32_t patch);
void NbreIrListFunc(int isc, void *holder, const char *ir_name_list);
void NbreIrVersionsFunc(int isc, void *holder, const char *ir_versions);
void NbreNrHandlerFunc(int isc, void *holder, const char *nr_handler);
void NbreNrResultFunc(int isc, void *holder, const char *nr_result);
void NbreDipRewardFunc(int isc, void *holder, const char *dip_reward);

void NbreVersionFunc_cgo(int isc, void *holder, uint32_t major, uint32_t minor,uint32_t patch) {
	NbreVersionFunc(isc, holder, major, minor, patch);
};

void NbreIrListFunc_cgo(int isc, void *holder, const char *ir_name_list) {
	NbreIrListFunc(isc, holder, ir_name_list);
};

void NbreIrVersionsFunc_cgo(int isc, void *holder, const char *ir_versions) {
	NbreIrVersionsFunc(isc, holder, ir_versions);
};

void NbreNrHandlerFunc_cgo(int isc, void *holder, const char *nr_handler) {
	NbreNrHandlerFunc(isc, holder, nr_handler);
};

void NbreNrResultFunc_cgo(int isc, void *holder, const char *nr_result) {
	NbreNrResultFunc(isc, holder, nr_result);
};

void NbreDipRewardFunc_cgo(int isc, void *holder, const char *dip_reward) {
	NbreDipRewardFunc(isc, holder, dip_reward);
};

*/
import "C"

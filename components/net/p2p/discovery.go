// Copyright (C) 2017 go-nebulas authors
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

package p2p

import (
	"context"
	"time"
)

/*
node can discover other node or can be discovered by another node
and then update the routing table.
*/
//TODO discorver other node
func (node *Node) Discovery(ctx context.Context) {

	//FIXME  the sync routing table rate can be dynamic
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			node.syncRoutingTable()
		case <-ctx.Done():
			log.Info("discovery service halting")
			return
		}
	}
}

//TODO sync route table
func (node *Node) syncRoutingTable() {

}

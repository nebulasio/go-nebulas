package main

import (
	"flag"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/nebulasio/go-nebulas/net"
)

type nodeIdx []int

// Len returns the size of nodeIdx
func (idx nodeIdx) Len() int {
	return len(idx)
}

// Swap swaps the ith with jth
func (idx nodeIdx) Swap(i, j int) {
	idx[i], idx[j] = idx[j], idx[i]
}

// Less returns true if ith <= jth else false
func (idx nodeIdx) Less(i, j int) bool {
	return idx[i] <= idx[j]
}

//
var (
	NodeCount     = flag.Int("node_count", 1000, "node count in network, default is 1000")
	NeighborCount = flag.Int64("neighbor_count", 50, "neighbor count in route table, default is 50")
	MaxTTL        = flag.Int64("max_ttl", 3, "max ttl, default is 3")
	LoopTimes     = flag.Int("loop_times", 20, "number of loop times, default is 20")
)

// Node the simulation of the node
type Node struct {
	id       int
	name     string
	neighbor []int
	bingo    bool
	ttl      int
}

func main() {
	flag.Parse()
	total := 0
	fmt.Printf("Usage: [-node_count] [-neighbor_count] [-max_ttl] [-loop_times]\n")
	for i := 0; i < *LoopTimes; i++ {
		count := gotask()
		total += count
	}

	fmt.Println("The average rate of coverage：", float32(total)/(float32(*LoopTimes*(*NodeCount))))

}

func gotask() int {

	nodeCount := int(*NodeCount)
	var nodes []*Node
	nodes = initRouteTable(nodeCount, nodes)

	random := rand.Intn(nodeCount)
	node := nodes[random]
	node.bingo = true
	broadcast(node, nodes)

	count := 0
	for _, v := range nodes {
		if v.bingo == true {
			count++
		}
	}
	fmt.Println("rate of coverage：", float32(count)/float32(nodeCount))
	return count
}

func initRouteTable(nodeCount int, nodes []*Node) []*Node {
	seed := newNode(0)
	nodes = append(nodes, seed)

	for i := 1; i < nodeCount; i++ {
		node := newNode(i)
		nodes = append(nodes, node)
		syncRoute(node, seed, nodes)
	}

	for k := 0; k < 10; k++ {
		for i := 0; i < nodeCount; i++ {
			node := nodes[i]
			rand.Seed(time.Now().UnixNano())
			randomList := rand.Perm(len(node.neighbor) - 1)

			for i := 0; i < int(math.Sqrt(float64(len(node.neighbor)))); i++ {
				id := node.neighbor[randomList[i]]
				tar := nodes[id]
				syncRoute(node, tar, nodes)
			}

		}
	}
	return nodes
}

func newNode(id int) *Node {
	networkKey, _ := net.GenerateEd25519Key()
	name, _ := peer.IDFromPublicKey(networkKey.GetPublic())
	node := &Node{
		id:       id,
		name:     name.Pretty(),
		neighbor: []int{},
		bingo:    false,
		ttl:      0,
	}
	return node
}

func broadcast(node *Node, nodes []*Node) {
	maxTTL := int(*MaxTTL)
	if node.ttl <= maxTTL {
		for _, v := range node.neighbor {
			n := nodes[v]
			if n.id != node.id {
				n.bingo = true
				if node.ttl <= maxTTL && n.ttl <= maxTTL {
					n.ttl = node.ttl + 1
					broadcast(n, nodes)
				}
			}
		}
	}
	return

}

func syncRoute(node *Node, target *Node, nodes []*Node) {

	neighborCount := int(*NeighborCount)
	if len(target.neighbor) < neighborCount {
		for id := range target.neighbor {
			node.neighbor = addNewNode(node.neighbor, id, neighborCount)
		}
		node.neighbor = addNewNode(node.neighbor, target.id, neighborCount)

		for id := range target.neighbor {
			n := nodes[id]
			n.neighbor = addNewNode(n.neighbor, node.id, neighborCount)
		}
		target.neighbor = addNewNode(target.neighbor, node.id, neighborCount)
		return
	}

	// target.neighbor = shuffle(target.neighbor)
	ret := getNearestNode(node, target.neighbor, nodes)
	for _, retID := range ret {
		node.neighbor = addNewNode(node.neighbor, int(retID), neighborCount)
		tempnode := nodes[int(retID)]
		tempnode.neighbor = addNewNode(tempnode.neighbor, node.id, neighborCount)
	}

	target.neighbor = addNewNode(target.neighbor, node.id, neighborCount)
	return
}

func getNearestNode(node *Node, ids []int, nodes []*Node) nodeIdx {
	var ret nodeIdx
	var tmp nodeIdx
	var tmpMap = make(map[int]int)
	// fmt.Println("nearest id:", len(ids))
	for _, id := range ids {
		tempnode := nodes[id]
		nodeNameInt := int(hash(node.name))
		tempnodeNameInt := int(hash(tempnode.name))
		distance := nodeNameInt ^ tempnodeNameInt
		// distance := node.id ^ id
		tmp = append(tmp, distance)
		sort.Sort(tmp)
		tmpMap[distance] = id
		if len(tmp) > len(ids)/2 {
			delete(tmpMap, tmp[len(tmp)-1])
			tmp = tmp[:len(tmp)-1]
		}
	}
	for _, v := range tmpMap {
		ret = append(ret, v)
	}

	return ret
}

func hash(str string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(str))
	return h.Sum32()
}

func addNewNode(ids []int, id int, limit int) []int {
	if len(ids) >= limit {
		count := len(ids) - limit
		ids = shuffle(ids)
		ids = ids[count+1:]
	}
	if !inArray(id, ids) {
		ids = append(ids, id)
	}
	return ids
}

func inArray(obj interface{}, array interface{}) bool {
	arrayValue := reflect.ValueOf(array)
	if reflect.TypeOf(array).Kind() == reflect.Array || reflect.TypeOf(array).Kind() == reflect.Slice {
		for i := 0; i < arrayValue.Len(); i++ {
			if arrayValue.Index(i).Interface() == obj {
				return true
			}
		}
	}
	return false
}

func shuffle(vals []int) []int {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]int, len(vals))
	perm := r.Perm(len(vals))
	for i, randIndex := range perm {
		ret[i] = vals[randIndex]
	}
	return ret
}

func generateRandomNumber(start int, end int, count int) []int {
	if end < start || (end-start) < count {
		return nil
	}

	nums := make([]int, 0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		num := r.Intn((end - start)) + start

		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}

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

package dag

import (
	"errors"
	"flag"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDispatcher_Start1(t *testing.T) {
	flag.Set("v", "true")
	flag.Set("log_dir", "/tmp")
	flag.Set("v", "3")
	flag.Parse()

	/*
				 1				16
				/ \				/ \
			   2   3		   17 18
			  /	  /  \			   \
			 4	 5	  6 		   19
			/	/ \	  / \
		   7   8   9 10 11
		   			\/
					12
					/ \
				   13 14
				   /
				  15
	*/
	dag := NewDag()

	dag.AddNode("1")
	dag.AddNode("2")
	dag.AddNode("3")
	dag.AddNode("4")
	dag.AddNode("5")
	dag.AddNode("6")
	dag.AddNode("7")
	dag.AddNode("8")
	dag.AddNode("9")
	dag.AddNode("10")
	dag.AddNode("11")
	dag.AddNode("12")
	dag.AddNode("13")
	dag.AddNode("14")
	dag.AddNode("15")
	dag.AddNode("16")
	dag.AddNode("17")
	dag.AddNode("18")
	dag.AddNode("19")
	// Add the edges (Note that given vertices must exist before adding an
	// edge between them)
	dag.AddEdge("1", "2")
	dag.AddEdge("1", "3")
	dag.AddEdge("2", "4")
	dag.AddEdge("3", "5")
	dag.AddEdge("3", "6")
	dag.AddEdge("4", "7")
	dag.AddEdge("5", "8")
	dag.AddEdge("5", "9")
	dag.AddEdge("6", "10")
	dag.AddEdge("6", "11")
	dag.AddEdge("9", "12")
	dag.AddEdge("10", "12")
	dag.AddEdge("12", "13")
	dag.AddEdge("13", "15")
	dag.AddEdge("12", "14")

	dag.AddEdge("15", "8")

	dag.AddEdge("16", "17")
	dag.AddEdge("16", "18")
	dag.AddEdge("18", "19")

	//dag.AddEdge("19", "16")

	runtime.GOMAXPROCS(runtime.NumCPU())

	msg, err := dag.ToProto()
	assert.Nil(t, err)

	dag1 := NewDag()
	dag1.FromProto(msg)

	txs := make([]string, 2)
	txs[0] = "a"
	txs[1] = "b"

	fmt.Println("runtime.NumCPU():", runtime.NumCPU())
	dp := NewDispatcher(dag, runtime.NumCPU(), 0, txs, func(node *Node, a interface{}) error {
		fmt.Println("dag Dispatcher key:", node.key, node.index)

		if node.key == 12 {
			fmt.Println(a)
			time.Sleep(time.Millisecond * 300)
			//return errors.New("test")
			return nil
		}
		time.Sleep(time.Millisecond * 100)
		return nil
	})

	err = dp.Run()
	assert.Nil(t, err)

	dp1 := NewDispatcher(dag, runtime.NumCPU(), 1, txs, func(node *Node, a interface{}) error {
		fmt.Println("key:", node.key, node.index)

		if node.key == 12 {
			fmt.Println(a)
			time.Sleep(time.Millisecond * 300)
			//return errors.New("test")
			return nil
		}
		time.Sleep(time.Millisecond * 100)
		return nil
	})

	err = dp1.Run()
	assert.NotNil(t, err)

	dag2 := NewDag()
	dag2.AddNode("1")

	fmt.Println("runtime.NumCPU():", runtime.NumCPU())
	dp2 := NewDispatcher(dag2, 8, 0, txs, func(node *Node, a interface{}) error {
		fmt.Println("key:", node.key, node.index)
		return nil
	})

	err = dp2.Run()
	assert.Nil(t, err)

	dag.AddEdge("19", "16")

	dp3 := NewDispatcher(dag, 8, 0, txs, func(node *Node, a interface{}) error {
		return errors.New("test error")
	})

	err = dp3.Run()
	assert.NotNil(t, err)
}

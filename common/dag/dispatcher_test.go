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
	"flag"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDispatcher_Start(t *testing.T) {
	type fields struct {
		concurrency int
		muTask      sync.Mutex
		dag         *Dag
		quitCh      chan bool
		queueCh     chan *Node
		tasks       map[string]*Task
		cursor      int
	}
	tests := []struct {
		name   string
		fields fields
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := &Dispatcher{
				concurrency: tt.fields.concurrency,
				muTask:      tt.fields.muTask,
				dag:         tt.fields.dag,
				quitCh:      tt.fields.quitCh,
				queueCh:     tt.fields.queueCh,
				tasks:       tt.fields.tasks,
				cursor:      tt.fields.cursor,
			}
			dp.Run()
		})
	}
}

func TestDispatcher_Start1(t *testing.T) {
	flag.Set("v", "true")
	flag.Set("log_dir", "/tmp")
	flag.Set("v", "3")
	flag.Parse()

	dag := NewDag()

	dag.AddNode("1", nil)
	dag.AddNode("2", nil)
	dag.AddNode("3", nil)
	dag.AddNode("4", nil)
	dag.AddNode("5", nil)
	dag.AddNode("6", nil)
	dag.AddNode("7", nil)
	dag.AddNode("8", nil)
	dag.AddNode("9", nil)
	dag.AddNode("10", nil)
	dag.AddNode("11", nil)
	dag.AddNode("12", nil)
	dag.AddNode("13", nil)
	dag.AddNode("14", nil)
	dag.AddNode("15", nil)
	dag.AddNode("16", nil)
	dag.AddNode("17", nil)
	dag.AddNode("18", nil)
	dag.AddNode("19", nil)
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
	dag.AddEdge("16", "17")
	dag.AddEdge("16", "18")
	dag.AddEdge("18", "19")

	runtime.GOMAXPROCS(runtime.NumCPU())

	msg, err := dag.ToProto()
	assert.Nil(t, err)

	dag1 := NewDag()
	dag1.FromProto(msg)

	fmt.Println("runtime.NumCPU():", runtime.NumCPU())
	dp := NewDispatcher(dag, runtime.NumCPU(), func(node *Node) error {
		fmt.Println("key:", node.Key)
		if node.Key == "12" {
			time.Sleep(time.Millisecond * 3000)
			//return errors.New("test")
			return nil
		}
		time.Sleep(time.Millisecond * 1000)
		return nil
	})

	err = dp.Run()
	assert.Nil(t, err)
}

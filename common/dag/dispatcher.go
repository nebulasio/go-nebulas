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
	"sync"
	"time"

	"github.com/nebulasio/go-nebulas/util/logging"
)

//type value interface{}

// Callback func node
type Callback func(*Node, interface{}) error

// Task struct
type Task struct {
	dependence int
	node       *Node
}

// Errors
var (
	ErrDagHasCirclular = errors.New("dag hava circlular")
)

// Dispatcher struct a message dispatcher dag.
type Dispatcher struct {
	concurrency int
	cb          Callback
	muTask      sync.Mutex
	dag         *Dag
	elapseInMs  int64
	quitCh      chan bool
	queueCh     chan *Node
	taskCounter int
	tasks       map[interface{}]*Task
	cursor      int
	err         error
	context     interface{}
}

// NewDispatcher create Dag Dispatcher instance.
func NewDispatcher(dag *Dag, concurrency int, elapseInMs int64, context interface{}, cb Callback) *Dispatcher {
	dp := &Dispatcher{
		concurrency: concurrency,
		elapseInMs:  elapseInMs,
		dag:         dag,
		cb:          cb,
		tasks:       make(map[interface{}]*Task, 0),
		taskCounter: 0,
		quitCh:      make(chan bool, 2*concurrency),
		queueCh:     make(chan *Node, 10240),
		cursor:      0,
		context:     context,
	}
	return dp
}

// Run dag dispatch goroutine.
func (dp *Dispatcher) Run() error {
	logging.VLog().Info("Starting Dag Dispatcher...")

	vertices := dp.dag.GetNodes()

	rootCounter := 0
	for _, node := range vertices {
		task := &Task{
			dependence: node.ParentCounter,
			node:       node,
		}
		task.dependence = node.ParentCounter
		dp.tasks[node.Key] = task

		if task.dependence == 0 {
			rootCounter++
			dp.push(node)
		}
	}
	if rootCounter == 0 && len(vertices) > 0 {
		dp.err = ErrDagHasCirclular
		return dp.err
	}

	dp.loop()

	return dp.err
}

// loop
func (dp *Dispatcher) loop() {
	logging.VLog().Debug("loop Dag Dispatcher.")

	//timerChan := time.NewTicker(time.Second).C

	if dp.dag.Len() < dp.concurrency {
		dp.concurrency = dp.dag.Len()
	}
	if dp.concurrency == 0 {
		return
	}
	wg := new(sync.WaitGroup)
	wg.Add(dp.concurrency)

	for i := 0; i < dp.concurrency; i++ {
		//logging.CLog().Info("loop Dag Dispatcher i:", i)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-dp.quitCh:
					logging.VLog().Debug("Stoped Dag Dispatcher.")
					return
				case msg := <-dp.queueCh:
					// callback todo
					err := dp.cb(msg, dp.context)

					if err != nil {
						dp.err = err
						dp.Stop()
					} else {
						err := dp.CompleteParentTask(msg)
						if err != nil {
							dp.err = err
							dp.Stop()
						}
					}
				}
			}
		}()
	}

	if dp.elapseInMs > 0 {
		go func() {
			timerChan := time.NewTimer(time.Duration(dp.elapseInMs) * time.Millisecond)

			<-timerChan.C
			if dp.cursor != dp.dag.Len() {
				dp.err = errors.New("dag timeout")
				dp.Stop()
			}
			return
		}()
	}

	wg.Wait()
}

// Stop stop goroutine.
func (dp *Dispatcher) Stop() {
	logging.VLog().Debug("Stopping dag Dispatcher...")

	for i := 0; i < dp.concurrency; i++ {
		select {
		case dp.quitCh <- true:
		default:
		}
	}
}

// push queue channel
func (dp *Dispatcher) push(vertx *Node) {
	dp.taskCounter++
	dp.queueCh <- vertx
}

// CompleteParentTask completed parent tasks
func (dp *Dispatcher) CompleteParentTask(node *Node) error {
	dp.muTask.Lock()
	defer dp.muTask.Unlock()

	key := node.Key

	vertices := dp.dag.GetChildrenNodes(key)
	for _, node := range vertices {
		err := dp.updateDependenceTask(node.Key)
		if err != nil {
			return err
		}
	}

	dp.cursor++

	if dp.cursor == dp.taskCounter {
		if dp.taskCounter < dp.dag.Len() {
			return ErrDagHasCirclular
		}
		dp.Stop()
	}

	return nil
}

// updateDependenceTask task counter
func (dp *Dispatcher) updateDependenceTask(key interface{}) error {
	if _, ok := dp.tasks[key]; ok {
		dp.tasks[key].dependence--
		//fmt.Println("Key:", key, " dependence:", dp.tasks[key].dependence)
		if dp.tasks[key].dependence == 0 {
			dp.push(dp.tasks[key].node)
		}
		if dp.tasks[key].dependence < 0 {
			return ErrDagHasCirclular
		}
	}
	return nil
}

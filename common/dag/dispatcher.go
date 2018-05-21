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
	"github.com/sirupsen/logrus"
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
	ErrTimeout         = errors.New("dispatcher execute timeout")
)

// Dispatcher struct a message dispatcher dag.
type Dispatcher struct {
	concurrency      int
	cb               Callback
	muTask           sync.Mutex
	dag              *Dag
	elapseInMs       int64
	quitCh           chan bool
	queueCh          chan *Node
	tasks            map[interface{}]*Task
	queueCounter     int
	completedCounter int
	isFinsih         bool
	finishCH         chan bool
	context          interface{}
}

// NewDispatcher create Dag Dispatcher instance.
func NewDispatcher(dag *Dag, concurrency int, elapseInMs int64, context interface{}, cb Callback) *Dispatcher {
	dp := &Dispatcher{
		concurrency:      concurrency,
		elapseInMs:       elapseInMs,
		dag:              dag,
		cb:               cb,
		tasks:            make(map[interface{}]*Task),
		queueCounter:     0,
		quitCh:           make(chan bool, concurrency),
		queueCh:          make(chan *Node, dag.Len()),
		completedCounter: 0,
		finishCH:         make(chan bool, 1),
		isFinsih:         false,
		context:          context,
	}
	return dp
}

// Run dag dispatch goroutine.
func (dp *Dispatcher) Run() error {
	logging.VLog().Debug("Starting Dag Dispatcher...")

	vertices := dp.dag.GetNodes()

	rootCounter := 0
	for _, node := range vertices {
		task := &Task{
			dependence: node.parentCounter,
			node:       node,
		}
		task.dependence = node.parentCounter
		dp.tasks[node.key] = task

		if task.dependence == 0 {
			rootCounter++
			dp.push(node)
		}
	}

	if rootCounter == 0 && len(vertices) > 0 {
		return ErrDagHasCirclular
	}

	return dp.execute()
}

// execute callback
func (dp *Dispatcher) execute() error {
	logging.VLog().Debug("loop Dag Dispatcher.")

	//timerChan := time.NewTicker(time.Second).C

	if dp.dag.Len() < dp.concurrency {
		dp.concurrency = dp.dag.Len()
	}
	if dp.concurrency == 0 {
		return nil
	}

	var err error
	go func() {
		for i := 0; i < dp.concurrency; i++ {
			go func() {
				for {
					select {
					case <-dp.quitCh:
						logging.VLog().Debug("Stoped Dag Dispatcher.")
						return
					case msg := <-dp.queueCh:
						err = dp.cb(msg, dp.context)

						if err != nil {
							dp.Stop()
						} else {
							isFinish, err := dp.onCompleteParentTask(msg)
							if err != nil {
								logging.VLog().WithFields(logrus.Fields{
									"err": err,
								}).Debug("Stoped Dag Dispatcher.")
								dp.Stop()
							}
							if isFinish {
								dp.Stop()
							}
						}
					}
				}
			}()
		}

		if dp.elapseInMs > 0 {
			deadlineTimer := time.NewTimer(time.Duration(dp.elapseInMs) * time.Millisecond)
			<-deadlineTimer.C
			err = ErrTimeout
			dp.Stop()
		}
	}()

	<-dp.finishCH
	return err
}

// Stop stop goroutine.
func (dp *Dispatcher) Stop() {
	logging.VLog().Debug("Stopping dag Dispatcher...")
	dp.muTask.Lock()
	defer dp.muTask.Unlock()
	if dp.isFinsih {
		return
	}
	dp.isFinsih = true

	for i := 0; i < dp.concurrency; i++ {
		select {
		case dp.quitCh <- true:
		default:
		}
	}
	dp.finishCH <- true
}

// push queue channel
func (dp *Dispatcher) push(vertx *Node) {
	dp.queueCounter++
	dp.queueCh <- vertx
}

// CompleteParentTask completed parent tasks
func (dp *Dispatcher) onCompleteParentTask(node *Node) (bool, error) {
	dp.muTask.Lock()
	defer dp.muTask.Unlock()

	key := node.key

	vertices := dp.dag.GetChildrenNodes(key)
	for _, node := range vertices {
		err := dp.updateDependenceTask(node.key)
		if err != nil {
			return false, err
		}
	}

	dp.completedCounter++

	if dp.completedCounter == dp.queueCounter {
		if dp.queueCounter < dp.dag.Len() {
			return false, ErrDagHasCirclular
		}
		return true, nil
	}

	return false, nil
}

// updateDependenceTask task counter
func (dp *Dispatcher) updateDependenceTask(key interface{}) error {
	if _, ok := dp.tasks[key]; ok {
		dp.tasks[key].dependence--
		if dp.tasks[key].dependence == 0 {
			dp.push(dp.tasks[key].node)
		}
		if dp.tasks[key].dependence < 0 {
			return ErrDagHasCirclular
		}
	}
	return nil
}

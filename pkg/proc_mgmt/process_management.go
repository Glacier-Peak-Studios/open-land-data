package proc_mgmt

import (
	"fmt"
	"log"
)

type ProcessState int

const (
	ProcessState_Running ProcessState = iota
	ProcessState_Finished
	ProcessState_Stopped
	ProcessState_Failed
	ProcessState_Unknown
)

type ProcessExecutor interface {
	Run()
	Value() *ProcessExecutable
	// String() string
	// Stop() error
	// Status() ProcessState
}

type ProcessExecutable struct {
	Name string
	Args []string
	Run  func()
}

// type ProcessExecutor interface{}

type OpenlandTask struct {
	Name    string
	Pid     int
	Status  string
	State   ProcessState
	Handler *ProcessExecutor
}

func NewOpenlandTask(name string, handler ProcessExecutor) *OpenlandTask {
	return &OpenlandTask{
		Name:    name,
		Handler: &handler,
	}
}

type OpenlandTaskChain struct {
	Name        string
	Pid         int
	Tasks       []*OpenlandTask
	chainLocked bool
}

func NewOpenlandTaskChain(name string) *OpenlandTaskChain {
	return &OpenlandTaskChain{
		Name:        name,
		Tasks:       []*OpenlandTask{},
		chainLocked: false,
	}
}

func (chain *OpenlandTaskChain) AddTask(task *OpenlandTask) {
	if !chain.chainLocked {
		chain.Tasks = append(chain.Tasks, task)
	}
}

func (chain *OpenlandTaskChain) Lock() {
	chain.chainLocked = true
}

// basic queue implementation with enqueue and dequeue methods
type ExecutionQueue struct {
	toExecute []*OpenlandTaskChain
}

func (q *ExecutionQueue) Enqueue(process *OpenlandTaskChain) {
	q.toExecute = append(q.toExecute, process)
}

func (q *ExecutionQueue) Dequeue() *OpenlandTaskChain {
	// remove the first process from the queue
	process := q.toExecute[0]
	q.toExecute = q.toExecute[1:]
	return process
}

func (q *ExecutionQueue) Len() int {
	return len(q.toExecute)
}

func (q *ExecutionQueue) Items() []*OpenlandTaskChain {
	return q.toExecute
}

type ProcessManager struct {
	Processes           map[int]*OpenlandTaskChain
	Queue               *ExecutionQueue
	ConcurrentProcesses int
	ProcessCount        int
	paused              bool
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		Processes:           make(map[int]*OpenlandTaskChain),
		Queue:               &ExecutionQueue{},
		ConcurrentProcesses: 1,
		ProcessCount:        0,
		paused:              false,
	}
}

func (pm *ProcessManager) Pause() {
	pm.paused = true
}

func (pm *ProcessManager) Resume() {
	pm.paused = false
}

func (pm *ProcessManager) AddProcess(name string, handler ProcessExecutor) {
	// generate a unique id for the process
	pid := len(pm.Processes) + 1
	// create process hash from name and pid
	task := &OpenlandTask{
		Name:    name,
		Handler: &handler,
		Pid:     pid,
	}

	pm.Processes[pid] = &OpenlandTaskChain{
		Name:  name,
		Pid:   pid,
		Tasks: []*OpenlandTask{task},
	}

}

func (pm *ProcessManager) QueueTaskChain(task *OpenlandTaskChain) {
	pm.Queue.Enqueue(task)
}

// func (pm *ProcessManager) RemoveProcess(name string) {
// 	// get process and stop it before deleting it
// 	process := pm.Processes[name]
// 	process.Handler.Stop()
// 	delete(pm.Processes, name)
// }

func (pm *ProcessManager) RunTaskChain(pid int) {
	// pm.ProcessCount++
	// get process and start it
	process := pm.Processes[pid]
	log.Printf("Starting chain %s", process.Name)
	for _, task := range process.Tasks {

		task.State = ProcessState_Running
		(*task.Handler).Run()
		task.State = ProcessState_Finished
	}
	delete(pm.Processes, pid)
	pm.ProcessCount--
}

// func (pm *ProcessManager) StopProcess(name string) {
// 	// get process and stop it
// 	process := pm.Processes[name]
// 	process.Handler.Stop()
// }

func (pm *ProcessManager) GetProcess(pid int) (*OpenlandTaskChain, error) {
	if _, ok := pm.Processes[pid]; ok {
		return pm.Processes[pid], nil
	}
	return nil, fmt.Errorf("process with id %d not found", pid)
}

func (pm *ProcessManager) GetProcessQueue() *ExecutionQueue {
	return pm.Queue
}

func (pm *ProcessManager) AddProcessToQueue(process *OpenlandTaskChain) {
	pm.Queue.Enqueue(process)
}

// get process and add it to queue

func (pm *ProcessManager) Start() {
	for {
		if !pm.paused && pm.Queue.Len() > 0 && pm.ProcessCount < pm.ConcurrentProcesses {
			process := pm.Queue.Dequeue()
			pm.Processes[process.Pid] = process
			pm.ProcessCount++
			go pm.RunTaskChain(process.Pid)
		}
	}
}

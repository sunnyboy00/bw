package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func doTask() {
	//耗时炒作(模拟)
	time.Sleep(200 * time.Millisecond)
	fmt.Println("aa")
	wg.Done()
}

//这里模拟的http接口,每次请求抽象为一个job
func handle() {
	//wg.Add(1)
	job := Job{}
	JobQueue <- job
}

var (
	MaxWorker = 1000
	MaxQueue  = 200000
	wg        sync.WaitGroup
)

type Worker struct {
	quit chan bool
}

func NewWorker() Worker {
	return Worker{
		quit: make(chan bool),
	}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w Worker) Start() {
	go func() {
		for {
			select {
			case <-JobQueue:
				// we have received a work request.
				doTask()
			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type Job struct {
}

var JobQueue chan Job = make(chan Job, MaxQueue)

type Dispatcher struct {
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < MaxWorker; i++ {
		worker := NewWorker()
		worker.Start()
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	d := NewDispatcher()
	d.Run()
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		handle()
	}
	wg.Wait()
}

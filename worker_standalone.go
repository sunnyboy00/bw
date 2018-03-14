package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	maxWorkers = 10
	maxJobNums = 100
)

type job struct {
	name     string
	duration time.Duration
}

func doWork(id int, j job) {
	fmt.Printf("协程%d: 开始 %s, 休眠 %fs\n", id, j.name, j.duration.Seconds())
	time.Sleep(j.duration)
	fmt.Printf("协程%d: 完成 %s!\n", id, j.name)
}

func main() {
	// channel for jobs
	jobs := make(chan job, 20)

	// start workers
	wg := &sync.WaitGroup{}
	wg.Add(maxWorkers)
	for i := 1; i <= maxWorkers; i++ {
		go func(i int) {
			defer wg.Done()

			for j := range jobs {
				doWork(i, j)
			}
		}(i)
	}

	// add jobs
	for i := 0; i < maxJobNums; i++ {
		name := fmt.Sprintf("任务%d", i)
		duration := time.Duration(rand.Intn(10)) * time.Millisecond
		fmt.Printf("添加任务: %s %s\n", name, duration)
		jobs <- job{name, duration}
	}
	close(jobs)

	// wait for workers to complete
	wg.Wait()
}

package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	//生成client 参数为默认
	client = &http.Client{}
	//生成要访问的url
	url = "http://127.0.0.1:8080/work"
)

const (
	maxWorkers = 10
	maxJobNums = 100000
)

type job struct {
	name     string
	duration time.Duration
}

func doWork(id int, j job) {
	fmt.Printf("协程%d: 开始 %s, 休眠 %fs\n", id, j.name, j.duration.Seconds())
	//time.Sleep(j.duration)

	// 提交请求
	param := fmt.Sprintf("name=myss%d-%s&delay=1s", id, j.name)
	req, err := http.NewRequest("POST", url, strings.NewReader(param))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 处理返回结果
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// 返回的状态码
	status := resp.StatusCode
	fmt.Println(status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))

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

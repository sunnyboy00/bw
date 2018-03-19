package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	MaxWorker = 10
	MaxQueues = 200
)

var (
	reg        *regexp.Regexp
	db         *sql.DB
	MinIDQueue chan int64 = make(chan int64)
	quit       chan bool  = make(chan bool)
	//生成client 参数为默认
	client = &http.Client{}
	//生成要访问的url
	url = "http://127.0.0.1:8080/work"
)

type Job struct {
	Id       int64
	Openid   string
	Nickname string
}

func init() {
	db, _ = sql.Open("mysql", "zzdcuser:40702f506be@tcp(localhost:3306)/dbzzdc?charset=utf8")
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.Ping()

	/*st, _ := db.Prepare("UPDATE tb_member SET fcword='' WHERE fcword!=''")
	defer st.Close()
	st.Exec()*/
}

func GetInfoList(minID int64) []Job {
	jobList := make([]Job, 0)

	sql := "SELECT id,openid,nickname FROM tb_member WHERE id>? ORDER BY id ASC LIMIT 100"
	stmt, e := db.Prepare(sql)
	if e != nil {
		fmt.Println(e.Error())
		return jobList
	}
	defer stmt.Close()

	rows, err := stmt.Query(minID)
	if err != nil {
		fmt.Println(err.Error())
		return jobList
	}
	defer rows.Close()

	lastId := minID
	for rows.Next() {
		var gs Job
		rows.Scan(&gs.Id, &gs.Openid, &gs.Nickname)
		jobList = append(jobList, gs)
		lastId = gs.Id
	}

	if err = rows.Err(); err != nil {
		fmt.Println(err.Error())
	}

	if lastId > minID {
		go func() {
			MinIDQueue <- lastId
		}()
	}

	if len(jobList) == 0 {
		go func() {
			//fmt.Println("waitting for exit..")
			quit <- true
			close(quit)
		}()
	}

	return jobList
}

func timeCost(start time.Time) {
	terminal := time.Since(start)
	fmt.Println("运行时长:", terminal)
}

func addFenciLogs(dpid, openid int64, words []string) {
	param := strings.Join(words, ",")
	fmt.Printf("%d => %s\n", dpid, param)
	st, _ := db.Prepare("UPDATE tb_member SET fcword=? WHERE id=?")
	defer st.Close()
	st.Exec(param, dpid)

	/*st, _ = db.Prepare("UPDATE tb_member SET nicknameent=REPLACE(nicknameent,?,'***') WHERE id=?")
	for _, v := range words {
		fmt.Printf("%d => %s\n", dpid, v)
		st.Exec(v, dpid)
	}*/
}

func doWork(i int, job Job) {
	//fmt.Printf("%#v\n", job)
	fmt.Printf("%d\t%s\t%s\t%d\n", job.Id, strings.TrimSpace(job.Openid), strings.TrimSpace(strings.Replace(job.Nickname, "\n", "", -1)), i)

	b, err := json.Marshal(job)
	//fmt.Printf("%#v\n", j)
	if err != nil {
		fmt.Println("json err:", err)
		return
	}
	param := bytes.NewBuffer([]byte(b))
	req, err := http.NewRequest("POST", url, param)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-type", "application/json; charset=UTF-8")

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
	/*st, _ := db.Prepare("INSERT INTO aa(id,openid) VALUES(?,?)")
	defer st.Close()
	st.Exec(job.id, strings.TrimSpace(job.openid))*/
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	defer timeCost(time.Now())

	// channel for jobs
	jobs := make(chan Job, MaxQueues)

	// start workers
	wg := &sync.WaitGroup{}
	for i := 1; i <= MaxWorker; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			for j := range jobs {
				doWork(i, j)
			}
		}(i)
	}

	go func() {
		for {
			select {
			case minID := <-MinIDQueue:
				jobList := GetInfoList(minID)
				for _, job := range jobList {
					jobs <- job
				}
			case <-time.After(200 * time.Millisecond):
				//fmt.Println("超时!")
				time.Sleep(200 * time.Millisecond)
			case <-quit:
				//fmt.Println("none")
				close(jobs)
				return
			}
		}
	}()

	go func() {
		MinIDQueue <- 0
	}()

	// wait for workers to complete
	wg.Wait()
	//select {}
}

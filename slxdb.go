package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	db *sql.DB
)

type Job struct {
	Id       int64  `json:"id"`
	Openid   string `json:"openid"`
	Nickname string `json:"nickname"`
}

func doWork(id int, j Job) {
	//fmt.Printf("worker%d: started %s, working for %f seconds\n", id, j.name, j.duration.Seconds())
	//time.Sleep(j.duration)
	//fmt.Printf("worker%d: completed %s!\n", id, j.name)
	//fmt.Printf("ok => %#v\n", j)
	st, _ := db.Prepare("INSERT INTO aa(id,openid,nickname) VALUES(?,?,?)")
	defer st.Close()
	st.Exec(j.Id, strings.TrimSpace(j.Openid), strings.TrimSpace(j.Nickname))
}

func requestHandler(jobs chan Job, w http.ResponseWriter, r *http.Request) {
	// Make sure we can only be called with an HTTP POST request.
	if r.Method != "POST" {
		//w.Header().Set("Allow", "POST")
		//w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "仅接受POST请求", http.StatusMethodNotAllowed)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)
	var job Job
	if err := json.Unmarshal(body, &job); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create Job and push the work onto the jobCh.
	//job := job{name, duration}
	//fmt.Printf("%#v\n", job)
	go func() {
		//fmt.Printf("added: %s %s\n", job.name, job.duration)
		jobs <- job
	}()

	// Render success.
	w.WriteHeader(http.StatusCreated)
	return
}

func init() {
	db, _ = sql.Open("mysql", "zzdcuser:40702f506be@tcp(localhost:3306)/dbzzdc?charset=utf8")
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.Ping()
}

func main() {
	var (
		maxQueueSize = flag.Int("max_queue_size", 100, "The size of job queue")
		maxWorkers   = flag.Int("max_workers", 5, "The number of workers to start")
		port         = flag.String("port", ":8080", "The server port")
	)
	flag.Parse()

	// create job channel
	jobs := make(chan Job, *maxQueueSize)

	// create workers
	for i := 1; i <= *maxWorkers; i++ {
		go func(i int) {
			for j := range jobs {
				doWork(i, j)
			}
		}(i)
	}

	// handler for adding jobs
	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		requestHandler(jobs, w, r)
	})
	log.Fatal(http.ListenAndServe(*port, nil))
}

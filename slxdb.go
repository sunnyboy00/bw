/*
CREATE TABLE aa (
  `id` int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '用户id',
  `openid` varchar(50) NOT NULL DEFAULT '' COMMENT '微信openid',
  `username` varchar(100) NOT NULL DEFAULT '' COMMENT '用户名',
  `nickname` varchar(100) NOT NULL DEFAULT '' COMMENT '昵称',
  `truename` varchar(100) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `mobile` varchar(20) NOT NULL DEFAULT '' COMMENT '手机号'
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
*/
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
	tb string
)

type Job struct {
	Id       int64  `json:"id"`
	Openid   string `json:"openid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Truename string `json:"truename"`
	Mobile   string `json:"mobile"`
}

func init() {
	db, _ = sql.Open("mysql", "zzdcuser:40702f506be@tcp(localhost:3306)/dbzzdc?charset=utf8")
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.Ping()

	tb = "aa"
}

func doWork(id int, j Job) {
	uid := 0
	e := db.QueryRow("SELECT id FROM "+tb+" WHERE id=? LIMIT 1", j.Id).Scan(&uid)
	if e != nil {
		st, _ := db.Prepare("INSERT INTO " + tb + "(id,openid,username,nickname,truename,mobile) VALUES(?,?,?,?,?,?)")
		defer st.Close()
		st.Exec(j.Id, strings.TrimSpace(j.Openid), strings.TrimSpace(j.Username), strings.TrimSpace(j.Nickname), strings.TrimSpace(j.Truename), strings.TrimSpace(j.Mobile))
	} else {
		st, _ := db.Prepare("UPDATE " + tb + " SET openid=?,username=?,nickname=?,truename=?,mobile=? WHERE id=?")
		defer st.Close()
		st.Exec(strings.TrimSpace(j.Openid), strings.TrimSpace(j.Username), strings.TrimSpace(j.Nickname), strings.TrimSpace(j.Truename), strings.TrimSpace(j.Mobile), uid)
	}
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

func main() {
	var (
		maxQueueSize = flag.Int("max_queue_size", 100, "The size of job queue")
		maxWorkers   = flag.Int("max_workers", 5, "The number of workers to start")
		port         = flag.String("port", ":8888", "The server port")
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

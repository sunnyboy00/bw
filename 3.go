package main

import (
	"fmt"
	"net/http"
	//"strconv"
	"sync"
)

var i int
var mutex = &sync.Mutex{}

func myfunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	//w.Header().Set("Content-Type", "application/zip")
	mutex.Lock()
	i++
	mutex.Unlock()
	//fmt.Fprintf(w, fmt.Sprintf("计数器：%s", strconv.Itoa(i)))
	w.Write([]byte(fmt.Sprintf("afdas: %d<br>", i)))
}

func main() {
	http.HandleFunc("/", myfunc)
	http.ListenAndServe(":8888", nil)
}

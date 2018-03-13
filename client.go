package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"net/http"
	"encoding/json"
)

type PayloadCollection struct {
	/*WindowsVersion string    `json:"version"`
	Token          string    `json:"token"`*/
	Payloads       []Payload `json:"data"`
}

type Payload struct {
	// [redacted]
	StorageFolder string
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) //限制同时运行的goroutines数量

	var s PayloadCollection
	var p Payload

	for i:=0; i<100000; i++ {
		//p.StorageFolder = fmt.Sprintf("GuangzhouVPN_%d", i)
		p.StorageFolder = fmt.Sprintf("VPN_%d", i)
		s.Payloads = append(s.Payloads, p)
	}

	b, err := json.Marshal(s)
	if err != nil {
		fmt.Println("json err:", err)
	}

	body := bytes.NewBuffer([]byte(b))
	res, err := http.Post("http://127.0.0.1:8888/", "application/json;charset=utf-8", body)
	if err != nil {
		log.Fatal(err)
		return
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Printf("%s", result)
}

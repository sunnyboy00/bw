package main

import (
	"compress/gzip"
	"fmt"
	//"io"
	"strings"
	"io/ioutil"
	"net/http"
)

func main() {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://count.taobao.com/counter6?keys=TCART_234_4e230c6031ce3f23105ccc071c9efc4c_q&t=1521081053187&_ksTS=1521081053189_22&callback=jsonp23", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	res, _ := client.Do(req)
	var body string
	if res.StatusCode == 200 {
		bodyByte, _ := ioutil.ReadAll(res.Body)
		body = string(bodyByte)
		fmt.Println(res.Header)

		switch res.Header.Get("Content-Encoding") {
		case "gzip":
			fmt.Println("-----------------------------")
			s := strings.NewReader(strings.TrimSpace(body))
			sor := ioutil.NopCloser(s)
			defer sor.Close()

			r, err := gzip.NewReader(sor)
			if err != nil {
				fmt.Println("aa")
				return
			}
			defer r.Close()

			respBody, err := ioutil.ReadAll(r)
			if err != nil {
				fmt.Println("bb")
				return
			}
			body = strings.TrimSpace(string(respBody))
			fmt.Println("-----------------------------")
		}
	}

	fmt.Println(body)
}

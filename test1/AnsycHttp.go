package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	start := time.Now()
	ch := make(chan string)
	urls := [...]string{"https://www.httpbin.or/delay/1", "https://www.httpbin.org/delay/2",
		"https://www.httpbin.org/delay/3", "https://www.httpbin.org/delay/4"}
	for _, url := range urls {
		go MakeRequest(url, ch)
	}
	for range urls {
		fmt.Println(<-ch)
	}
	fmt.Printf("all spend:%.2fs \n", time.Since(start).Seconds())
}

func MakeRequest(url string, ch chan<- string) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	secs := time.Since(start).Seconds()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	resp.Body.Close()
	ch <- fmt.Sprintf("spend time:%.2f,body: %s %s", secs, body, url)
}

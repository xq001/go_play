package main

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/asmcos/requests"
	"strings"
)

func main() {
	req := requests.Requests()
	//req.Header.Set("Content-Type", "application/json")
	//无论代理是http还是https都要使用http开头
	//req.Proxy("http://116.196.66.197:9999")
	req.SetTimeout(10)
	resp, _ := req.Get("https://studygolang.com/articles/11590")
	fmt.Println(resp.Text())
	s := resp.Text()
	//var json map[string]interface{}
	//resp.Json(json)
	//fmt.Println(json["origin"])
	defer resp.R.Body.Close()
	doc, err := htmlquery.Parse(strings.NewReader(s))
	if err != nil {
		fmt.Println(err)
	}
	list := htmlquery.InnerText(htmlquery.FindOne(doc, "//*[@class=\"dark\"]"))
	fmt.Println(list)
}

//func main() {
//	proxy := func(_ *http.Request) (*url.URL, error) {
//		return url.Parse("http://47.92.98.68:3128")
//	}
//	p,_:=url.Parse("http://47.92.98.68:3128")
//	fmt.Println(*p)
//
//	tr := &http.Transport{
//		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
//		Proxy:           proxy,
//	}
//
//	client := &http.Client{Transport: tr}
//	seedUrl := "https://httpbin.org/get"
//	req, err := http.NewRequest("GET", seedUrl, nil)
//	req.Header.Add("User-Agent",
//		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
//	resp, err := client.Do(req)
//	if err != nil {
//		fmt.Println(err)
//		panic(err)
//	}
//
//	defer resp.Body.Close()
//
//	body, _ := ioutil.ReadAll(resp.Body)
//	fmt.Printf("%s\n", body)
//}

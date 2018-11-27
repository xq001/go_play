package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/go-ini/ini"
	"github.com/kataras/iris"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var PublishUrl string
var VerifyUrl string

func init() {

	config, err := ini.Load("./conf_for_iris.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	PublishUrl = config.Section("BASE").Key("publish_url").String()
	VerifyUrl = config.Section("BASE").Key("verify_url").String()

}
func main() {
	app := iris.Default()
	app.Get("/hello", Hello)
	app.Get("/get_proxy", GetProxyHandle)
	app.Get("/ping", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"message": "pong",
		})
	})

	// listen and serve on http://0.0.0.0:8080.
	app.Run(iris.Addr(PublishUrl))
}
func GetProxyHandle(ctx iris.Context) {

	proxyList := VerifyProxy(GetProxyIpList())
	//fmt.Println("验证后ip:", proxyList)
	ctx.JSON(iris.Map{
		"ip_list": proxyList,
	})
}

func Hello(ctx iris.Context) {
	ctx.Writef("hello iris")
}
func GetProxyIpList() []string {
	//f, err := os.Open("./ne.html")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//defer f.Close()
	//bfRd := bufio.NewReader(f)
	client := http.Client{}
	//提交请求
	reqest, err := http.NewRequest("GET", "http://www.xicidaili.com/wn/1", nil)
	if err != nil {
		fmt.Println(err)
	}
	reqest.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
	response, err := client.Do(reqest)
	if err != nil {
		fmt.Println(err)
	}
	doc, err := htmlquery.Parse(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	//table:=htmlquery.FindOne(doc,"//tr[0]")

	var proxyList []string
	for _, i := range htmlquery.Find(doc, "//*[@id=\"ip_list\"]/tbody/tr")[1:] {
		td := htmlquery.Find(i, "//td")
		ip := htmlquery.InnerText(td[1])
		port := htmlquery.InnerText(td[2])
		protocol := strings.ToLower(htmlquery.InnerText(td[5]))
		proxyList = append(proxyList, protocol+"://"+ip+":"+port)
	}

	//fmt.Println("待验证ip个数", len(proxyList))
	return proxyList

}
func VerifyProxy(urlList []string) []string {
	ch := make(chan string)
	var res []string
	urls := urlList
	for _, u := range urls {
		go VerifyOneProxy(u, VerifyUrl, ch)
	}
	for range urls {
		i := <-ch
		if i != "" {
			res = append(res, i)
		}

	}
	return res
}
func VerifyOneProxy(verifyProxyUrl string, verifyUrl string, ch chan<- string) {
	//验证一个代理
	urli := url.URL{}
	//go语言代理无论http或https都需要写成http
	proxyUrl, _ := urli.Parse(strings.Replace(verifyProxyUrl, "https", "http", 1))
	client := http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyUrl),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Duration(5 * time.Second),
	}

	//提交请求
	reqest, err := http.NewRequest("GET", verifyUrl, nil)
	if err != nil {
		ch <- fmt.Sprint("")
		return
	}
	reqest.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
	//处理返回结果
	response, err := client.Do(reqest)
	if err != nil {
		//fmt.Println(err)
		ch <- fmt.Sprint("")
		return
	}

	body, _ := ioutil.ReadAll(response.Body)
	//fmt.Println(string(body))
	var myjson []map[string]interface{}
	errjson := json.Unmarshal(body, &myjson)
	if errjson != nil || len(myjson) == 0 {
		//fmt.Println(errjson)
		ch <- fmt.Sprint("")
		return
	}
	defer response.Body.Close()
	ch <- fmt.Sprint(verifyProxyUrl)
}

package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func main() {
	//VerifyProxy(prosy_list)
	//ll := []string{"http://202.112.237.102:3128","http://202.112.237.112:3128"}
	proxyList := VerifyProxy(getProxyIpList())
	fmt.Println("验证后ip:", proxyList)
}
func getProxyIpList() []string {
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

	fmt.Println("待验证ip个数", len(proxyList))
	return proxyList

}
func VerifyProxy(urlList []string) []string {
	ch := make(chan string)
	var res []string
	urls := urlList
	for _, u := range urls {
		go VerifyOneProxy(u, "https://maimai.cn/contact/interest_contact/eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1IjoxMTk4MjA4NDAsImxldmVsIjoyLCJ0IjoiY3R0In0.fnz6vNCb63n2j-Frr6H_vu1LuG1jgfoq2oPOITSAJdA", ch)
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
	proxy_url, _ := urli.Parse(strings.Replace(verifyProxyUrl, "https", "http", 1))
	client := http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxy_url),
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

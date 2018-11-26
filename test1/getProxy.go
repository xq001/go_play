package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	//prosy_list := getProxyIpList()
	//VerifyProxy(prosy_list)
	ll := []string{"http://202.112.237.102:3128"}
	VerifyProxy(ll)
}
func getProxyIpList() []string {
	f, err := os.Open("./ne.html")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	bfRd := bufio.NewReader(f)
	doc, err := htmlquery.Parse(bfRd)
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

	fmt.Println(proxyList)
	return proxyList

}
func VerifyProxy(urlList []string) {
	ch := make(chan string)
	urls := urlList
	for _, u := range urls {
		go VerifyOneProxy(u, "https://www.httpbin.org/get", ch)
	}
	for range urls {
		fmt.Println(<-ch)
	}
	//return []string{}
}
func VerifyOneProxy(verifyProxyUrl string, verifyUrl string, ch chan<- string) {
	//验证一个代理
	urli := url.URL{}
	proxy_url, _ := urli.Parse(verifyProxyUrl)
	c := http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxy_url),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Duration(5 * time.Second),
	}

	resp, err := c.Get(verifyUrl)
	if err != nil {
		fmt.Println(err)
		ch <- fmt.Sprint(nil)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	var myjson map[string]interface{}
	if err := json.Unmarshal(body, &myjson); err != nil {
		fmt.Println(err)
		ch <- fmt.Sprint(nil)
		return
	}
	defer resp.Body.Close()
	ch <- fmt.Sprint(verifyProxyUrl)
}

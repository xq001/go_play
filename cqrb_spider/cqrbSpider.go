package main

import (
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/astaxie/beego/logs"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/go-ini/ini"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var userAgentList = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 8.0; Pixel 2 Build/OPD3.170816.012) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1",
	"Mozilla/5.0 (iPad; CPU OS 11_0 like Mac OS X) AppleWebKit/604.1.34 (KHTML, like Gecko) Version/11.0 Mobile/15A5341f Safari/604.1"}
var session *mgo.Session
var database *mgo.Collection
var logger *logs.BeeLogger

func init() {
	//初始化配置文件
	config, err := ini.Load("./conf.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	mongodbUrl := config.Section("mongodb").Key("URL").String()
	databaseName := config.Section("mongodb").Key("DATABASE").String()
	tableName := config.Section("mongodb").Key("CONNECTION").String()
	logfilePath := config.Section("log").Key("logfile").String()
	//fmt.Println(databaseName,tableName)
	//初始化日志
	logger = logs.NewLogger()
	logger.SetLogger("file", `{"filename":"`+logfilePath+`"}`)
	//初始化数据库连接
	_session, err := mgo.Dial(mongodbUrl) //传入数据库的地址，可以传入多个，具体请看接口文档
	if err != nil {
		panic(err)
	}
	session = _session
	//defer session.Close() //用完记得关闭
	//dbcon := session.DB("test").C("test")
	session.SetMode(mgo.Monotonic, true)
	database = session.DB(databaseName).C(tableName)

}

type NewsItem struct {
	NewsTitle string
	NewsText  string
	NewsUrl   string
}

type Item struct {
	Version string
	Title   string
	News    []*NewsItem
}
type AllItem struct {
	Time    string
	AllNews []*Item
}

var allItem = new(AllItem)

func main() {

	allItem.Time = time.Now().Format("2006-01-02")
	logger.Info("############开始爬取今日重庆日报数据##########")
	logger.Info(time.Now().String())
	scrapy("http://epaper.cqrb.cn/html/cqrb/2018-12/13/001/node.htm")
	logger.Info("存储结果")
	if err := InsertToDB(allItem); err != nil {
		logger.Error(fmt.Sprint(err))
	}
	defer session.Close()
	logger.Info(time.Now().String())
	logger.Info("############结束爬取##########")
}

func scrapy(url string) {
	body := getNews(url)
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		fmt.Println(err)
		//logger.Error(fmt.Sprint(err))
	}
	item := new(Item)
	temp := htmlquery.FindOne(doc, "/html/body/table/tbody/tr[1]/td[1]/table/tbody/tr[1]/td/table[2]/tbody/tr/td[1]")
	item.Version = htmlquery.InnerText(temp)
	item.Title = htmlquery.InnerText(htmlquery.FindOne(temp, "./strong"))

	for _, oneNew := range htmlquery.Find(doc, "/html/body/table/tbody/tr[1]/td[1]/table/tbody/tr[2]/td/table/tbody/tr") {
		news := new(NewsItem)
		news.NewsTitle = htmlquery.InnerText(htmlquery.FindOne(oneNew, "./td[2]/a/div"))
		news.NewsUrl = "http://epaper.cqrb.cn" + htmlquery.InnerText(htmlquery.FindOne(oneNew, "./td[2]/a/@href"))
		news.NewsText = getOneNews(news.NewsUrl)
		item.News = append(item.News, news)
	}
	//data, err := json.Marshal(item)
	//fmt.Println(string(data))
	//存储数据
	allItem.AllNews = append(allItem.AllNews, item)

	nextUrl := ""
	_len := len(htmlquery.Find(doc, "/html/body/table/tbody/tr[1]/td[1]/table/tbody/tr[1]/td/table[2]/tbody/tr/td[2]/table/tbody/tr/td[2]/a"))
	if _len == 1 {
		if htmlquery.InnerText(htmlquery.FindOne(doc, "/html/body/table/tbody/tr[1]/td[1]/table/tbody/tr[1]/td/table[2]/tbody/tr/td[2]/table/tbody/tr/td[2]/a")) == "下一版" {
			nextUrl = "http://epaper.cqrb.cn" + htmlquery.InnerText(htmlquery.FindOne(doc, "/html/body/table/tbody/tr[1]/td[1]/table/tbody/tr[1]/td/table[2]/tbody/tr/td[2]/table/tbody/tr/td[2]/a/@href"))

		}
	} else if _len == 2 {
		nextUrl = "http://epaper.cqrb.cn" + htmlquery.InnerText(htmlquery.FindOne(doc, "/html/body/table/tbody/tr[1]/td[1]/table/tbody/tr[1]/td/table[2]/tbody/tr/td[2]/table/tbody/tr/td[2]/a[2]/@href"))
	}
	if nextUrl != "" {
		scrapy(nextUrl)
	}
}
func randChoiceList(list []string) string {
	/*随机取出列表中的一个数*/
	rand.Seed(time.Now().Unix())
	length := len(list)
	return list[rand.Intn(length)]
}

func getNews(url string) string {
	client := http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		//logger.Error(fmt.Sprint(err))
	}
	request.Header.Set("User-Agent", randChoiceList(userAgentList))
	response, err := client.Do(request)

	if err != nil {
		logger.Error(fmt.Sprint(err))
	}
	body, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()
	return string(body)

}
func getOneNews(url string) string {
	client := http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(fmt.Sprint(err))
	}
	request.Header.Set("User-Agent", randChoiceList(userAgentList))
	response, err := client.Do(request)

	if err != nil {
		logger.Error(fmt.Sprint(err))
	}
	body, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	doc, err := htmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		logger.Error(fmt.Sprint(err))
	}
	return htmlquery.InnerText(htmlquery.FindOne(doc, "//*[@id=\"ozoom\"]"))

}

func InsertToDB(item *AllItem) error {
	//data := AllItem{}
	count, _ := database.Find(bson.M{"time": time.Now().Format("2006-01-02")}).Count()
	if count == 0 {
		if err := database.Insert(item); err != nil {
			logger.Error(fmt.Sprint(err))
			return err
		}
		return nil
	}
	logger.Warn("插入的数据已经存在!")
	return errors.New("插入数据异常")

}

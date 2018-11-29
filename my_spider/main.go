package main

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/astaxie/beego/logs"
	"github.com/globalsign/mgo"
	"github.com/go-ini/ini"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"os"
	"regexp"
	"strings"
	"time"
)

type Item struct {
	//ID          bson.ObjectId `bson:"_id"`
	BidTitle    string `bson:"bid_title"`
	ReleaseTime string `bson:"release_time"`
	BidArea     string `bson:"bid_area"`
	BidIndustry string `bson:"bid_industry"`
	BidUrl      string `bson:"bid_url"`
	BidText     string `bson:"bid_text"`
	//Bidder      string        `bson:"bidder"`
}

//var session *mgo.Session
var database *mgo.Collection
var config *ini.File
var mongodbUrl string
var databaseName string
var tableName string
var logfilePath string
var threadNum int
var formatUrl string

func init() {

	_cfg, err := ini.Load("./conf.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	config = _cfg
	mongodbUrl = config.Section("mongodb").Key("URL").String()
	databaseName = config.Section("mongodb").Key("DATABASE").String()
	tableName = config.Section("mongodb").Key("CONNECTION").String()
	logfilePath = config.Section("log").Key("logfile").String()
	threadNum, _ = config.Section("spider").Key("thread_num").Int()
	formatUrl = config.Section("spider").Key("format_url").String()
	session, err := mgo.Dial(mongodbUrl) //传入数据库的地址，可以传入多个，具体请看接口文档
	if err != nil {
		panic(err)
	}
	//defer session.Close() //用完记得关闭
	//dbcon := session.DB("test").C("test")
	session.SetMode(mgo.Monotonic, true)
	database = session.DB(databaseName).C(tableName)
}
func GetConfig() *ini.File {
	return config
}
func GetDataBase() *mgo.Collection {
	return database
}

func InsertToDB(item Item) error {
	con := GetDataBase()
	if err := con.Insert(item); err != nil {
		return err
	}
	return nil
}

//colly爬虫框架
func main() {
	cfg := GetConfig()
	logger := logs.NewLogger()
	logger.SetLogger("file", `{"filename":"`+logfilePath+`"}`)
	timestart := time.Now()
	var item Item
	re, _ := regexp.Compile("\\s")
	fmt.Println("开始采集...")
	logger.Info("###########################开始采集#######################")
	// Instantiate default collector
	c := colly.NewCollector(
		// Turn on asynchronous requests
		//colly.Async(true),
		// Attach a debugger to the collector
		//colly.Debugger(&debug.LogDebugger{}),
		//colly.AllowedDomains("www.chinabidding.org.cn"),
		// AllowURLRevisit instructs the Collector to allow multiple downloads of the same URL
		colly.AllowURLRevisit(),
	)
	// create a request queue with 2 consumer threads
	q, _ := queue.New(
		threadNum, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 200000}, // Use default queue storage
	)
	//设置请求超时时间，默认10秒
	c.SetRequestTimeout(20 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		//fmt.Println("Visiting", r.URL)
		r.Headers.Set("User-Agent",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
				"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
	})
	c.OnResponse(func(r *colly.Response) {
		//item.BidUrl = fmt.Sprint(r.Request.URL)
		//fmt.Println("OnResponse")
	})
	c.OnError(func(r *colly.Response, e error) {
		//打印错误，将超时URL重新放入队列
		logger.Error(fmt.Sprint(r.Request.URL, e))
		re, _ := regexp.Compile("(?i:timeout)")
		if re.MatchString(fmt.Sprint(e)) {
			q.AddURL(fmt.Sprint(r.Request.URL))
			logger.Warn("重新加入队列" + fmt.Sprint(r.Request.URL) + "当前队列大小:" + fmt.Sprint(q.Size()))
		}
		//fmt.Println("OnError:", r.Request.URL, e)
	})
	c.OnScraped(func(r *colly.Response) {
		//提取需要的信息，并存入数据库
		//fmt.Println("OnScraped:", )
		doc, _ := htmlquery.Parse(strings.NewReader(string(r.Body)))
		item.BidTitle = htmlquery.InnerText(htmlquery.FindOne(doc, "//*[@id=\"cphMain_tle\"]"))
		item.ReleaseTime = htmlquery.InnerText(htmlquery.FindOne(doc, "//*[@id=\"cphMain_tm\"]"))
		item.BidIndustry = re.ReplaceAllString(
			htmlquery.InnerText(htmlquery.FindOne(doc, "//*[@id=\"fld\"]")), "")
		item.BidArea = re.ReplaceAllString(
			htmlquery.InnerText(htmlquery.FindOne(doc, "//*[@id=\"dtr\"]")), "")
		item.BidText = htmlquery.OutputHTML(htmlquery.FindOne(doc, "//*[@id=\"dinfo\"]"), true)
		item.BidUrl = fmt.Sprint(r.Request.URL)
		if err := InsertToDB(item); err != nil {
			logger.Error(fmt.Sprint(err))
		}
		//jsons, errs := json.Marshal(item) //转换成JSON返回的是byte[]
		//if errs != nil {
		//	fmt.Println(errs.Error())
		//}
		//fmt.Println(string(jsons))
	})

	//c.Visit("http://www.chinabidding.org.cn/BidInfoDetails_bid_11039642.html")
	//c.Wait()
	// Start scraping in five threads on https://httpbin.org/delay/2
	//
	//for i := 11037602; i < 11039642; i++ {
	//	c.Visit(fmt.Sprintf("http://www.chinabidding.org.cn/BidInfoDetails_bid_%d.html", i))
	//}
	//// Wait until threads are finished
	//c.Wait()

	start, _ := cfg.Section("spider").Key("startid").Int()
	end, _ := cfg.Section("spider").Key("endid").Int()

	for i := start; i <= end; i++ {
		// Add URLs to the queue
		//q.AddURL(fmt.Sprintf("http://www.chinabidding.org.cn/BidInfoDetails_bid_%d.html", i))
		q.AddURL(fmt.Sprintf(formatUrl, i))
	}

	// Consume URLs
	q.Run(c)
	timeend := time.Now()
	fmt.Println("spend time:", timeend.Sub(timestart))

}

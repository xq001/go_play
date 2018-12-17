package main

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/astaxie/beego/logs"
	"github.com/globalsign/mgo"
	"github.com/go-ini/ini"
	"github.com/go-redis/redis"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

var userAgentList = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 8.0; Pixel 2 Build/OPD3.170816.012) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1",
	"Mozilla/5.0 (iPad; CPU OS 11_0 like Mac OS X) AppleWebKit/604.1.34 (KHTML, like Gecko) Version/11.0 Mobile/15A5341f Safari/604.1"}

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

var RedisDb *redis.ClusterClient

//var session *mgo.Session
var DataBase *mgo.Collection
var Config *ini.File
var ThreadNum int
var Logger *logs.BeeLogger

func init() {
	RedisDb = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"192.168.108.181:6379", "192.168.108.182:6379", "192.168.108.183:6379",
			"192.168.109.237:6379", "192.168.109.238:6379", "192.168.109.239:6379"},
	})

	_cfg, err := ini.Load("./conf.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	Config = _cfg
	mongodbUrl := Config.Section("mongodb").Key("URL").String()
	databaseName := Config.Section("mongodb").Key("DATABASE").String()
	tableName := Config.Section("mongodb").Key("CONNECTION").String()
	logfilePath := Config.Section("log").Key("logfile").String()
	ThreadNum, _ = Config.Section("spider").Key("thread_num").Int()
	fmt.Println(mongodbUrl, databaseName, tableName, logfilePath)

	session, err := mgo.Dial(mongodbUrl) //传入数据库的地址，可以传入多个，具体请看接口文档
	if err != nil {
		panic(err)
	}
	//defer session.Close() //用完记得关闭
	//dbcon := session.DB("test").C("test")
	session.SetMode(mgo.Monotonic, true)
	DataBase = session.DB(databaseName).C(tableName)

	Logger = logs.NewLogger()
	Logger.SetLogger("file", `{"filename":"`+logfilePath+`"}`)
}

func InsertToDB(item Item) error {
	if err := DataBase.Insert(item); err != nil {
		return err
	}
	return nil
}

func addUrl() {
	//将url放入redis
	for i := 9846295; i <= 11363841; i++ {
		RedisDb.SAdd("bid_2018_urls",
			fmt.Sprintf("http://www.chinabidding.org.cn/BidInfoDetails_bid_%d.html", i))
	}
	RedisDb.Close()
}

var urlKEY = "bid_2018_urls"

//var urlKEY = "xqset"

func main() {
	addUrl()
	//scrapy()

}
func scrapy() {
	fmt.Println("开始采集")
	timestart := time.Now()
	var item Item
	Logger.Info("###########################开始采集#######################")
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
		ThreadNum, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 20000}, // Use default queue storage
	)
	//设置请求超时时间，默认10秒
	c.SetRequestTimeout(20 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		//fmt.Println("Visiting", r.URL)
		r.Headers.Set("User-Agent", randChoiceList(userAgentList))
		//添加url
		qlen, _ := q.Size()
		if qlen < 10 {
			for range [100]string{} {
				url := RedisDb.SPop(urlKEY).Val()
				if url == "" {
					Logger.Error("redis列表为空")
					break
				}
				q.AddURL(url)
			}
		}
	})
	c.OnResponse(func(r *colly.Response) {
		//item.BidUrl = fmt.Sprint(r.Request.URL)
		//fmt.Println("OnResponse")
	})

	c.OnError(func(r *colly.Response, e error) {
		//打印错误，将超时URL重新放入队列
		//if r.StatusCode==404||r.StatusCode==500{
		//
		//}
		Logger.Error(fmt.Sprint(r.Request.URL, e))
		re, _ := regexp.Compile("(?i:timeout)")
		if re.MatchString(fmt.Sprint(e)) {
			q.AddURL(fmt.Sprint(r.Request.URL))
			Logger.Warn("重新加入队列" + fmt.Sprint(r.Request.URL) + "当前队列大小:" + fmt.Sprint(q.Size()))
		}
		//fmt.Println("OnError:", r.Request.URL, e)
	})
	re, _ := regexp.Compile("\\s")
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
		BidText := htmlquery.InnerText(htmlquery.FindOne(doc, "//*[@id=\"dinfo\"]"))
		item.BidText = strings.TrimSpace(strings.Replace(BidText, "\n", "", -1))
		//item.BidText = htmlquery.OutputHTML(htmlquery.FindOne(doc, "//*[@id=\"dinfo\"]"), true)
		item.BidUrl = fmt.Sprint(r.Request.URL)
		if err := InsertToDB(item); err != nil {
			Logger.Error(fmt.Sprint(err))
		}
		//jsons, errs := json.Marshal(item) //转换成JSON返回的是byte[]
		//if errs != nil {
		//	fmt.Println(errs.Error())
		//}
		//fmt.Println(string(jsons))

	})
	//先添加1个url
	url := RedisDb.SPop(urlKEY).Val()
	if url == "" {
		Logger.Error("redis列表为空")
		panic("bid_2018_urls 无数据")
	}
	q.AddURL(url)
	// Consume URLs
	q.Run(c)
	timeend := time.Now()
	fmt.Println("spend time:", timeend.Sub(timestart))
	fmt.Println("结束采集")

}

func randChoiceList(list []string) string {
	/*随机取出列表中的一个数*/
	rand.Seed(time.Now().Unix())
	length := len(list)
	return list[rand.Intn(length)]
}

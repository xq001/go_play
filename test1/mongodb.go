package main

import (
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type Item struct {
	ID          bson.ObjectId `bson:"_id"`
	BidTitle    string        `bson:"bid_title"`
	ReleaseTime string        `bson:"release_time"`
	Bidder      string        `bson:"bidder"`
	BidArea     string        `bson:"bid_area"`
	BidText     string        `bson:"bid_text"`
}

func main() {
	//var item Item
	session, err := mgo.Dial("mongodb://192.168.110.51:27017") //传入数据库的地址，可以传入多个，具体请看接口文档

	if err != nil {
		panic(err)
	}
	defer session.Close() //用完记得关闭
	c := session.DB("test").C("test")

	err = c.Insert(&Item{ID: bson.NewObjectId(),
		BidTitle: "这是一个测试", ReleaseTime: "2018年10月22日09:32:20",
		Bidder: "xxxx", BidArea: "重庆", BidText: "qweertewtwertwretwe"})
	if err != nil {
		fmt.Println(err)
	}
	//if err := c.Find(bson.M{}).One(&item); err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(item.BidTitle)
}

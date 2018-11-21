package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"reflect"
)

func main() {
	redisdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"192.168.108.181:6379", "192.168.108.182:6379", "192.168.108.183:6379",
			"192.168.109.237:6379", "192.168.109.238:6379", "192.168.109.239:6379"},
	})
	defer redisdb.Close()
	val, err := redisdb.LRange("maimai_list", 0, -1).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(len(val), reflect.TypeOf(val))
	//flag:=0
	//for _, value := range val {
	//
	//	val2, _:=redisdb.LRem("maimai_list", 0, value).Result()
	//	flag+=1
	//	fmt.Println(flag,val2)
	//}
	val2, err := redisdb.SAdd("maimai_set", val).Result()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(val2)
	//cluster, err := redis.NewCluster(
	//	&redis.Options{
	//		StartNodes:   []string{"192.168.108.181:6379", "192.168.108.182:6379", "192.168.108.183:6379"},
	//		ConnTimeout:  50 * time.Millisecond,
	//		ReadTimeout:  50 * time.Millisecond,
	//		WriteTimeout: 50 * time.Millisecond,
	//		KeepAlive:    16,
	//		AliveTime:    60 * time.Second,
	//	})
	//
	//if err != nil {
	//	log.Fatalf("redis.New error: %s", err.Error())
	//}
	//
	//name, err := redis.Strings(cluster.Do("keys", "na*"))
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(name)
	//defer cluster.Close()

}

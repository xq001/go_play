package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {

	name := "./testwritefile.txt"
	content := "Hello, xxbandy.github.io!\n"
	WriteWithBufio(name, content)

}

//使用os.OpenFile()相关函数打开文件对象，并使用文件对象的相关方法进行文件写入操作
func WriteWithFileWrite(name, content string) {
	fileObj, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Failed to open the file", err.Error())
		os.Exit(2)
	}
	defer fileObj.Close()

	for i := 0; i <= 10000000; i++ {
		if _, err := fileObj.WriteString(strconv.Itoa(i) + content); err != nil {
			fmt.Println("fail writing to the file with os.OpenFile and *File.WriteString method.", content)
		}
	}
	//contents := []byte(content)
	//if _,err := fileObj.Write(contents);err == nil {
	//	fmt.Println("Successful writing to thr file with os.OpenFile and *File.Write method.",content)
	//}
}

//使用bufio包中Writer对象的相关方法进行数据的写入
func WriteWithBufio(name, content string) {
	if fileObj, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0644); err == nil {
		defer fileObj.Close()
		writeObj := bufio.NewWriterSize(fileObj, 1024*1024*100)
		timeStart := time.Now()
		//
		//for i := 0; i <= 10000000; i++ {
		//	writeObj.WriteString(strconv.Itoa(i) + content)
		//}

		//使用Write方法,需要使用Writer对象的Flush方法将buffer中的数据刷到磁盘
		buf := []byte(content)
		for i := 0; i < 100000; i++ {
			writeObj.Write(buf)
		}
		writeObj.Flush()

		timeEnd := time.Now()
		fmt.Println("spend time:", timeEnd.Sub(timeStart))
	}
}

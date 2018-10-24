package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"time"
)

type MainController struct {
	beego.Controller
}

func (this *MainController) Get() {

	fmt.Println("sleep 12")
	time.Sleep(20 * time.Second)
	this.Ctx.WriteString("hello world")

}

func main() {

	beego.Router("/", &MainController{})

	beego.Run()

}

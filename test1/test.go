package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	var s = "asdasdtimeoutasdTimeout,TIMEOUT"
	r, _ := regexp.Compile("(?i:timeout)")
	fmt.Println(r.MatchString(s))
	fmt.Println(fmt.Sprint(strings.Contains(s, "timeout")))
}

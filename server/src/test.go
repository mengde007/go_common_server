package main

import (
	"fmt"
	mj "majiangserver"
)

func init() {
	testF = 5
	TestDic[1] = "2222"
}

var testF = 10

var TestDic = map[int]string{
	1: "1111"}

func main() {
	//fmt.Println("hello world", TestDic[1])

	fmt.Println("Start Test")

	mj.TestBao()

	fmt.Println("End Test")
}

package main

import (
	"fmt"

	"github.com/xecus/connectedcar/adapter"
	"github.com/xecus/connectedcar/config"
)

func main() {

}

func main2() {

	globalConfig, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	rc := adapter.NewRedisClient()
	rc.Init(globalConfig)

	err = rc.Write("testkey", "testval")
	if err != nil {
		panic(err)
	}

	value, err := rc.Read("testkey")
	if err != nil {
		panic(err)
	}
	fmt.Println("value=", value)
}

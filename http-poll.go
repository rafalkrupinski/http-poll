package main

import (
	"github.com/rafalkrupinski/http-poll/etsy"
	"flag"
	"fmt"
	ht "github.com/rafalkrupinski/revapigw/http"
	"net/http"
	"strconv"
	"github.com/rafalkrupinski/http-poll/tasks"
	"time"
	"net/url"
)

func onErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var shopId int
	flag.IntVar(&shopId, "s", 0, "ShopId")
	flag.Parse()

	srcUrl, err := url.Parse("http://openapi.etsy.com/v2/shops/" + strconv.Itoa(shopId) + "/receipts")
	onErr(err)

	proxyUrl, err := srcUrl.Parse("http://localhost:8080")
	onErr(err)

	spec := &tasks.TaskSpecification{
		SourceSpecification: &tasks.SourceSpecification{
			SourceAddress: srcUrl,
		},
		Frequency:     time.Second * 5,
		TargetAddress: "http://localhost:9090/receipt",
		InClient : ht.NewClientBuilder().WithTransport(&http.Transport{Proxy: http.ProxyURL(proxyUrl)}).Build(),
		OutClient : ht.NewClientBuilder().Build(),
		Task: etsy.New(),
	}

	task := tasks.NewTaskState(spec)

	start(task)
}

func start(task *tasks.TaskState) {
	ticker := time.NewTicker(task.Frequency)
	go doRun(task)
	for range ticker.C {
		go doRun(task)
	}
}

func doRun(task *tasks.TaskState){
	err := task.Run()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(task)
}

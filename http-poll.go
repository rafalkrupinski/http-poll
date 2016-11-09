package main

import (
	"github.com/rafalkrupinski/http-poll/tasks"
	ht "github.com/rafalkrupinski/revapigw/http"
	"net/http"
	"net/url"
	"time"
	"flag"
	"strconv"
)

func onErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	shopId := flag.Int("i", 0, "ShopId")
	flag.Parse()

	spec := &tasks.TaskSpecification{
		SourceAddress: "http://openapi.etsy.com/v2/shops/" + strconv.Itoa(*shopId) + "/receipts",
		IdLimitArg:    "max_created",
		SortOrder:     tasks.Desc,
		Frequency:     time.Second * 5,
		IdPath:        "$.results[*].creation_tsz+",
		TargetAddress: "http://localhost:9090/receipt",
	}

	proxyUrl, err := url.Parse("http://localhost:8080")
	onErr(err)
	inClient := ht.NewClientBuilder().WithTransport(&http.Transport{Proxy: http.ProxyURL(proxyUrl)}).Build()
	outClient := ht.NewClientBuilder().Build()

	task, err := tasks.NewTaskState(spec, inClient, outClient)
	onErr(err)

	ticker := time.NewTicker(spec.Frequency)

	task.Run()
	for range ticker.C {
		task.Run()
	}

}

package http_poll

import (
	"fmt"
	"github.com/rafalkrupinski/http-poll/tasks"
	"time"
)

func Start(task *tasks.TaskState) {
	ticker := time.NewTicker(task.Frequency)
	go doRun(task)
	for range ticker.C {
		go doRun(task)
	}
}

func doRun(task *tasks.TaskState) {
	err := task.Run()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(task)
}

package http_poll

import (
	"fmt"
	"github.com/rafalkrupinski/http-poll/tasks"
	"time"
	"os/signal"
	"os"
	"syscall"
)

func StartMulti(allTasks []*tasks.TaskState) {
	for _, task := range allTasks {
		go doLoop(task)
	}
}

func Start(task *tasks.TaskState) {
	go doLoop(task)
}

func Wait() {
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}

func doLoop(task *tasks.TaskState) {
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

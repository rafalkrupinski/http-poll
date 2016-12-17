package http_poll

import (
	"github.com/rafalkrupinski/http-poll/tasks"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

func StartMulti(allTasks []*tasks.TaskInst) {
	for _, task := range allTasks {
		go doLoop(task)
	}
}

func Start(task *tasks.TaskInst) {
	go doLoop(task)
}

func Wait() {
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}

func doLoop(task *tasks.TaskInst) {
	ticker := time.NewTicker(task.Spec.Frequency)
	go doRun(task)
	for range ticker.C {
		go doRun(task)
	}
}

func doRun(task *tasks.TaskInst) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error in %v %s", r, debug.Stack())
		}
	}()

	err := task.Run()
	if err != nil {
		log.Println(err)
	}
	log.Println(task)
}

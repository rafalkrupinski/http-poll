package tasks

import (
	"fmt"
	"github.com/dghubble/sling"
	"github.com/docker/libkv/store"
	"github.com/rafalkrupinski/http-poll/persist"
	ht "github.com/rafalkrupinski/rev-api-gw/http"
	"net/http"
	"net/url"
	"time"
)

type Task interface {
	SetStore(store.Store)
	Process(*http.Response) (next *url.URL, error error)
	OnSuccess()
}

type SourceSpecification struct {
	//URL of the data to poll
	SourceAddress *url.URL
}

type TaskSpecification struct {
	Name string

	*SourceSpecification

	//Poll frequency
	Frequency time.Duration

	TargetAddress string

	InClient  *http.Client
	OutClient *http.Client

	Task
}

type TaskState struct {
	*TaskSpecification
	next *url.URL
}

func NewTaskState(spec *TaskSpecification) *TaskState {
	ts := &TaskState{
		spec,
		nil,
	}

	if ts.Task != nil {
		ts.Task.SetStore(persist.GetPrefixed(spec.Name))
	}

	return ts
}

func (ts *TaskState) Run() error {
	nextUrl := ts.nextUrl()
	fmt.Println(nextUrl)

	//TODO handle URL with #fragment
	req, err := sling.New().Client(ts.InClient).Get(nextUrl.String()).Request()
	if err != nil {
		panic(err)
	}

	resp, err := ts.InClient.Do(req)
	if err != nil {
		return err
	}

	if ts.Task != nil {
		next, err := ts.Task.Process(resp)
		if err != nil {
			return err
		}

		if next != nil {
			ts.next = next
		}
	}

	err = ts.post(resp)

	if err == nil {
		ts.Task.OnSuccess()
	}

	return err
}

func (ts *TaskState) post(resp *http.Response) error {
	passReq, err := http.NewRequest(http.MethodPost, ts.TargetAddress, resp.Body)
	if err != nil {
		return err
	}

	if v := resp.Header.Get(ht.CONTENT_LEN); v != "" {
		passReq.Header.Add(ht.CONTENT_LEN, v)
	}

	_, err = ts.OutClient.Do(passReq)
	return err
	//TODO handle http errors
}

func (ts *TaskState) nextUrl() *url.URL {
	if ts.next == nil {
		return ts.SourceSpecification.SourceAddress
	} else {
		return ts.next
	}
}

func (t *TaskState) String() string {
	return fmt.Sprintf("%v %v", t.next, t.Task)
}

package tasks

import (
	"errors"
	"fmt"
	"github.com/dghubble/sling"
	"github.com/docker/libkv/store"
	"github.com/rafalkrupinski/http-poll/persist"
	ht "github.com/rafalkrupinski/rev-api-gw/http"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Processor interface {
	// Return a URL to the next batch of data based on passed default (or base) URL and Processor's state.
	Next(*url.URL) (*url.URL, error)

	// Get state of this processor so the tasks module can manage it's persistence
	State() ProcessorState

	// Update the state of this processor based on the response
	Process(*http.Response) error
}

type ProcessorState interface {
	Persist(store.Store) error
	Read(store.Store) error
}

type RemoteSpecification struct {
	//URL of the data to poll
	Address *url.URL

	//Method  string

	*http.Client
}

type TaskSpecification struct {
	Name string

	//Poll frequency
	Frequency time.Duration

	Source *RemoteSpecification

	ProcessorFactory

	Destination *RemoteSpecification
}

type ProcessorFactory func() Processor

type TaskInst struct {
	Spec *TaskSpecification
	p    Processor
	s    store.Store
}

func NewTaskInst(spec *TaskSpecification) *TaskInst {
	ts := &TaskInst{
		Spec: spec,
	}

	if spec.ProcessorFactory != nil {
		ts.p = spec.ProcessorFactory()
	} else {
		ts.p = new(defaultProcessor)
	}

	state := ts.p.State()
	if state != nil {
		ts.s = persist.GetPrefixed(spec.Name)
		state.Read(ts.s)
	}

	return ts
}

func (ts *TaskInst) Run() error {
	resp, err := ts.retrieve()
	if err != nil {
		return err
	}

	state := ts.p.State()

	err = ts.p.Process(resp)
	if err != nil {
		goto Error
	}

	err = ts.send(resp)
	if err != nil {
		goto Error
	}

	if state != nil {
		err = state.Persist(ts.s)
		if err != nil {
			goto Error
		}
	}

	return nil

	// Move processor to the previous state
Error:
	if state != nil {
		err = state.Read(ts.s)
	}

	return err
}

func (task *TaskInst) retrieve() (*http.Response, error) {
	nextUrl, err := task.nextUrl()
	if err != nil {
		return nil, err
	}

	log.Print(nextUrl)

	//TODO handle URL with #fragment
	req, err := sling.New().Client(task.Spec.Source.Client).Get(nextUrl.String()).Request()
	if err != nil {
		return nil, err
	}

	resp, err := task.Spec.Source.Client.Do(req)
	if err == nil && resp.StatusCode/100 != 2 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		} else {
			return nil, errors.New(string(body))
		}
	}
	return resp, err
}

func (task *TaskInst) send(resp *http.Response) error {
	passReq, err := http.NewRequest(http.MethodPost, task.Spec.Destination.Address.String(), resp.Body)
	if err != nil {
		return err
	}

	if v := resp.Header.Get(ht.CONTENT_LEN); v != "" {
		passReq.Header.Add(ht.CONTENT_LEN, v)
	}

	_, err = task.Spec.Destination.Client.Do(passReq)
	return err
	//TODO handle http errors
}

func (task *TaskInst) nextUrl() (*url.URL, error) {
	return task.p.Next(task.Spec.Source.Address)
}

func (t *TaskInst) String() string {
	return fmt.Sprintf("%v %v", t.Spec.Name, t.p)
}

type defaultProcessor struct{}

func (*defaultProcessor) Next(addr *url.URL) (*url.URL, error) {
	return addr, nil
}

func (p *defaultProcessor) State() ProcessorState {
	return p
}

func (*defaultProcessor) Process(*http.Response) error {
	return nil
}

func (*defaultProcessor) Persist(store.Store) error {
	return nil
}

func (*defaultProcessor) Read(store.Store) error {
	return nil
}

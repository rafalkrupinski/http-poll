package tasks

import (
	"bytes"
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
	// if spec.Init is defined, this method will be called with the result
	Init() error

	// Return a URL to the next batch of data based on passed default (or base) URL and Processor's state.
	Next(*url.URL) (*url.URL, error)

	// Get state of this processor so the tasks module can manage it's persistence
	State() ProcessorState

	// Update the state of this processor based on the response
	Process(*RemoteData) error
}

type ProcessorState interface {
	Persist(store.Store) error
	Read(store.Store) error
}

type RemoteSpecification struct {
	//URL of the data to poll
	Address string

	//Method  string

	Client *http.Client
}

type RemoteData struct {
	Body   []byte
	Header http.Header
}

type TaskSpecification struct {
	Name string

	Init *RemoteSpecification

	//Poll frequency
	Frequency,
	Delay time.Duration

	Source *RemoteSpecification

	ProcessorFactory

	Destination *RemoteSpecification
}

type ProcessorFactory func(*TaskSpecification) Processor

type TaskInst struct {
	Spec *TaskSpecification
	p    Processor
	s    store.Store
}

func NewTaskInst(spec *TaskSpecification) (*TaskInst, error) {
	ts := &TaskInst{
		Spec: spec,
	}

	if spec.ProcessorFactory != nil {
		ts.p = spec.ProcessorFactory(spec)
	} else {
		ts.p = new(defaultProcessor)
	}

	state := ts.p.State()
	if state != nil {
		ts.s = persist.GetPrefixed(spec.Name)
		state.Read(ts.s)
	}

	return ts, nil
}

func (t *TaskInst) Init() error {
	return t.p.Init()
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

	log.Print(err)

	if state != nil {
		err = state.Read(ts.s)
	}

	return err
}

func (task *TaskInst) retrieve() (*RemoteData, error) {
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
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if resp.StatusCode/100 != 2 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		} else {
			return nil, errors.New(string(body))
		}
	}

	data := &RemoteData{Body: body}

	return data, err
}

func (task *TaskInst) send(resp *RemoteData) error {
	passReq, err := http.NewRequest(http.MethodPost, task.Spec.Destination.Address, bytes.NewReader(resp.Body))
	if err != nil {
		return err
	}

	if _, ok := resp.Header[ht.CONTENT_LEN]; ok {
		v := resp.Header.Get(ht.CONTENT_LEN)
		passReq.Header.Add(ht.CONTENT_LEN, v)
	}

	passReq.Header.Set(ht.CONTENT_TYPE, "application/json")

	_, err = task.Spec.Destination.Client.Do(passReq)
	//TODO handle http errors
	return err
}

func (task *TaskInst) nextUrl() (*url.URL, error) {
	addr, err := url.Parse(task.Spec.Source.Address)
	if err != nil {
		return nil, err
	}
	return task.p.Next(addr)
}

func (t *TaskInst) String() string {
	return fmt.Sprintf("%v %v", t.Spec.Name, t.p)
}

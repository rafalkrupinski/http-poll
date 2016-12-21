package tasks

import (
	"github.com/docker/libkv/store"
	"github.com/rafalkrupinski/http-poll/persist"
	"net/url"
	"time"
)

type TimedTaskInst struct {
	nextTime,
	DefaultStartTime time.Time
	TimedTask
	StartTimeParam,
	EndTimeParam string
	TimeSpan time.Duration
}

func (t *TimedTaskInst) Init() error {
	init, err := t.TimedTask.Init()
	if err != nil {
		return err
	}
	if !init.IsZero() {
		t.nextTime = init
	} else {
		t.nextTime = t.DefaultStartTime
	}
	return nil
}

func (t *TimedTaskInst) Next(addr *url.URL) (*url.URL, error) {
	q := addr.Query()
	endTime := t.nextTime.Add(t.TimeSpan)
	t.addParams(&q, endTime)
	addr.RawQuery = q.Encode()

	t.nextTime = endTime
	return addr, nil
}

func (t *TimedTaskInst) addParams(v *url.Values, endTime time.Time) {
	v.Add(t.StartTimeParam, t.TimedTask.EncodeTime(t.nextTime))
	v.Add(t.EndTimeParam, t.TimedTask.EncodeTime(endTime))
}

func (t *TimedTaskInst) State() ProcessorState {
	return t
}

func (t *TimedTaskInst) Process(*RemoteData) error {
	return nil
}

// ProcessorState

const nextTime = "nextTime"

func (t *TimedTaskInst) Persist(s store.Store) error {
	typed := &persist.TypedStore{s}
	return typed.PutInt64(nextTime, t.nextTime.Unix())
}

func (t *TimedTaskInst) Read(s store.Store) error {
	typed := persist.TypedStore{s}
	var storedTime int64
	exists, err := typed.Int64(nextTime, &storedTime)
	if err != nil {
		return nil
	}

	if exists {
		t.nextTime = time.Unix(storedTime, 0)
	}
	return nil
}

// TimedTask

type TimedTask interface {
	Init() (time.Time, error)
	EncodeTime(time.Time) string
}

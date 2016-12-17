package tasks

import (
	"github.com/docker/libkv/store"
	"net/url"
)

type defaultProcessor struct{}

func (*defaultProcessor) Next(addr *url.URL) (*url.URL, error) {
	return addr, nil
}

func (p *defaultProcessor) State() ProcessorState {
	return p
}

func (*defaultProcessor) Process(*ResponseData) error {
	return nil
}

func (*defaultProcessor) Init(*TaskInst) error {
	return nil
}

func (*defaultProcessor) Persist(store.Store) error {
	return nil
}

func (*defaultProcessor) Read(store.Store) error {
	return nil
}

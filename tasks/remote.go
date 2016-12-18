package tasks

import (
	"github.com/dghubble/sling"
	"net/http"
	"log"
)

func (rs *RemoteSpecification) request(method string) (*sling.Sling, *http.Request, error) {
	client := sling.New().Client(rs.Client)
	switch method {
	case http.MethodGet:client.Get(rs.Address)
	case http.MethodPost:client.Post(rs.Address)
	default: log.Fatal("Unexpected method ", method)
	}

	return client, client.Request()
}

func (rs *RemoteSpecification) GetObject(target interface{}) (*http.Response, error) {
	client, req, err := rs.request(http.MethodGet)
	if req != nil {
		return nil, err
	}

	return client.Do(req, target, nil)
}


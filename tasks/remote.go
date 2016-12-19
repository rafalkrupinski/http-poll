package tasks

import (
	"github.com/dghubble/sling"
	"log"
	"net/http"
)

func (rs *RemoteSpecification) request(method string) (*sling.Sling, *http.Request, error) {
	client := sling.New().Client(rs.Client)
	switch method {
	case http.MethodGet:
		client.Get(rs.Address)
	case http.MethodPost:
		client.Post(rs.Address)
	default:
		log.Fatal("Unexpected method ", method)
	}

	req, err := client.Request()
	return client, req, err
}

func (rs *RemoteSpecification) GetObject(target interface{}) (*http.Response, error) {
	client, req, err := rs.request(http.MethodGet)
	if req != nil {
		return nil, err
	}

	return client.Do(req, target, nil)
}

package etsy

import (
	"errors"
	"fmt"
	"github.com/docker/libkv/store"
	"github.com/eleme/jsonpath"
	"github.com/rafalkrupinski/http-poll/uints64"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type etsyTask struct {
	//paging stop condition. stop paging if encountered this id
	limitId,

	// max ID from current paging session, to be used as a new limitId when the current one is found
	maxId uint64
}

func (task *etsyTask) Process(resp *http.Response) (*url.URL, error) {
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ids, nextPage, maxId, err := parseResponse(buf)
	if err != nil {
		return nil, err
	}

	task.maxId = uints64.Max(task.maxId, maxId)

	if len(ids) == 0 {
		return nil, errors.New("No results")
	}

	requestUrl := *resp.Request.URL
	nextUrl := requestUrl

	if uints64.UnsortedContains(ids, task.limitId) {
		nextPage = 0
		task.limitId = task.maxId
	}

	if nextPage == 0 {
		q := nextUrl.Query()
		q.Del("page")
		nextUrl.RawQuery = q.Encode()
	} else {
		q := nextUrl.Query()
		q.Set("page", uints64.Itoa(nextPage))
		nextUrl.RawQuery = q.Encode()
	}

	return &nextUrl, nil
}

func parseResponse(buf []byte) (ids []uint64, nextPage, maxId uint64, err error) {
	//TODO take receipt_id as ctor param
	idPath, err := jsonpath.ParsePath("$.results[*].receipt_id")
	if err != nil {
		return
	}

	pagePath, err := jsonpath.ParsePath("$.pagination.next_page")
	if err != nil {
		return
	}

	data, err := jsonpath.EvalPathsInBytes(buf, []*jsonpath.Path{idPath, pagePath})
	if err != nil {
		return
	}

	for {
		result, hasNext := data.Next()
		if !hasNext {
			break
		}

		if result.Value == nil {
			continue
		}

		number, err := strconv.ParseUint(string(result.Value), 10, 64)
		if err != nil {
			return []uint64{}, 0, 0, err
		}

		keys := result.Keys
		key := string(keys[len(keys)-1].([]byte))
		// TODO figure out way to read this const from the pattern
		if key == "receipt_id" {
			id := number
			newMaxId := uints64.Max(maxId, number)
			fmt.Printf("%d %d => %d\n", maxId, id, newMaxId)

			ids = append(ids, id)
			maxId = newMaxId
		} else {
			nextPage = number
			fmt.Println("Next page: ", nextPage)
		}
	}
	return
}

func New() *etsyTask {
	return new(etsyTask)
}

func (t *etsyTask) String() string {
	return fmt.Sprintf("%v %v", t.limitId, t.maxId)
}

func (t *etsyTask) OnSuccess() {
}

func (t *etsyTask) SetStore(db store.Store) {
}

package tasks

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	ht "github.com/rafalkrupinski/revapigw/http"
	"github.com/NodePrime/jsonpath"
	"github.com/dghubble/sling"
	"strconv"
	"strings"
	"time"
	"github.com/rafalkrupinski/http-poll/uints64"
	"net/url"
)

type TaskSpecification struct {
	//URL of the data to poll
	SourceAddress    string

	//param name to limit results (e.g. since)
	IdLimitArg       string

	//Poll frequency
	Frequency        time.Duration

	//JSONPath of the results' ID used to limit
	IdPath           string

	// URL of the resource, which will match IdPath and return Id of the last known object
	LastResultSource string

	// The order in which the service returns its results -> use either min or max Id in the next hit
	SortOrder        ISortOrder

	TargetAddress    string
}

type RuntimeSpec struct {
	*TaskSpecification

	idSelect maxFunc

	parseId  parseIdFunc

	nextUrl  nextUrl
}

func NewRuntimeSpec(ts*TaskSpecification) *RuntimeSpec {
	result := &RuntimeSpec{
		TaskSpecification:ts,
		idSelect : ts.SortOrder.Max,
		parseId : func(data[]byte) (uint64, error) {
			return strconv.ParseUint(string(data), 10, 64)
		},
	}

	return result
}

type TaskState struct {
	*RuntimeSpec
	lastId    uint64
	idPath    []*jsonpath.Path
	inClient  *http.Client
	outClient *http.Client
}

func NewTaskState(spec*TaskSpecification, inClient  *http.Client, outClient *http.Client) (*TaskState, error) {
	ts := &TaskState{
		RuntimeSpec:NewRuntimeSpec(spec),
		lastId:spec.SortOrder.InitialId(),
	}

	if ts.IdLimitArg != "" && ts.IdPath != "" {
		ts.nextUrl = ts.nextUrlWithId
		//ts.nextUrl = ts.nextUrlPlain

		paths, err := jsonpath.ParsePaths(ts.IdPath)
		if err != nil {
			return nil, err
		}
		ts.idPath = paths
	} else {
		ts.nextUrl = ts.nextUrlPlain
	}

	ts.inClient = inClient
	ts.outClient = outClient
	return ts, nil
}

func (ts*TaskState)Run() {
	nextUrl := ts.nextUrl()
	fmt.Println(nextUrl)
	req, err := sling.New().Get(nextUrl).Request()
	if err != nil {
		panic(err)
	}

	resp, err := ts.inClient.Do(req)
	if err != nil {
		panic(err)
	}

	buf, err := ioutil.ReadAll(resp.Body)

	data, err := jsonpath.EvalPathsInBytes(buf, ts.idPath)

	lastId := ts.lastId

	for {
		if result, ok := data.Next(); ok {
			id, err := ts.parseId(result.Value)
			if err != nil {
				fmt.Println(err)
			} else {
				newId := ts.idSelect(lastId, id)
				fmt.Printf("%d %d => %d\n", lastId, id, newId)
				lastId = newId
			}
		} else {
			break
		}
	}

	fmt.Println(lastId)
	ts.lastId = lastId

	if data.Error != nil {
		fmt.Println(data.Error)
		return
	}
	passReq, _ := http.NewRequest(http.MethodPost, ts.TargetAddress, bytes.NewReader(buf))

	if v := resp.Header.Get(ht.CONTENT_LEN); v != "" {
		passReq.Header.Add(ht.CONTENT_LEN, v)
	}

	ts.outClient.Do(passReq)

}

type nextUrl func() string

func (ts*TaskState) nextUrlWithId() string {
	addr, _ := url.Parse(ts.SourceAddress)
	query := addr.Query()
	query.Set(ts.IdLimitArg, strconv.FormatUint(ts.lastId, 10))
	addr.RawQuery = query.Encode()
	return addr.String()
}

func (ts*TaskState) nextUrlPlain() string {
	return ts.SourceAddress
}

type maxFunc func(uint64, uint64) uint64

type ISortOrder interface {
	Max(uint64, uint64) uint64
	InitialId() uint64
}

type SortOrder int

const (
	Desc SortOrder = -1
	Asc SortOrder = 1
)

func ParseSortOrder(v string) (SortOrder, error) {
	switch strings.ToUpper(v) {
	case "ASC":
		return Asc, nil
	case "DESC":
		return Desc, nil
	default:
		return 0, errors.New("Unupported SortOrder")
	}
}

func (o SortOrder) Max(a, b uint64) uint64 {
	if o == Asc {
		return uints64.Max(a, b)
	} else {
		return uints64.Min(a, b)
	}
}

func (o SortOrder)InitialId() uint64 {
	if o == Asc {
		return uints64.MIN
	} else {
		return uints64.MAX
	}
}

type parseIdFunc func([]byte) (uint64, error)


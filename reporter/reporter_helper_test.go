package reporter_test

import (
	"fmt"
	//"github.com/SpringerPE/graphite-du-report/reporter"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

type FakeDataFetcher struct {
	Responses map[string]*pb.MetricDetailsResponse
}

func NewFakeDataFetcher() *FakeDataFetcher {
	return &FakeDataFetcher{make(map[string]*pb.MetricDetailsResponse)}
}

func (fetcher *FakeDataFetcher) FetchData(url string) (*pb.MetricDetailsResponse, error) {
	if val, ok := fetcher.Responses[url]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("No data to fetch")
}

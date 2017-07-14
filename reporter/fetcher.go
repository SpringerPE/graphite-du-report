package reporter

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

type limiter chan struct{}

func (l limiter) enter() { l <- struct{}{} }
func (l limiter) leave() { <-l }

func newLimiter(l int) limiter {
	return make(chan struct{}, l)
}

var errTimeout = fmt.Errorf("Max tries exceeded")

type Fetcher interface {
	FetchData(url string) (*pb.MetricDetailsResponse, error)
}

type DataFetcher struct {
	http.Client
	Retries int
}

type metricDetailsFlat struct {
	*pb.MetricDetails
	Name string
}

type jsonMetricDetailsResponse struct {
	Metrics    []metricDetailsFlat
	FreeSpace  uint64
	TotalSpace uint64
}

func NewDataFetcher(timeout time.Duration, retries int) Fetcher {
	return &DataFetcher{Client: http.Client{Timeout: timeout * time.Second}, Retries: retries}
}

/*Fetches a data response from the graphite details endpoint
 */
func (fetcher *DataFetcher) FetchData(url string) (*pb.MetricDetailsResponse, error) {
	var metricsResponse pb.MetricDetailsResponse
	var response *http.Response
	var err error
	tries := 1

retry:
	if tries > fetcher.Retries {
		log.Println("Tries exceeded while trying to fetch data")
		return nil, errTimeout
	}
	response, err = fetcher.Get(url)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Println("Error during communication with client")
		tries++
		time.Sleep(300 * time.Millisecond)
		goto retry
	} else {

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println(err)
			log.Println("Error while reading client's response")
			tries++
			time.Sleep(300 * time.Millisecond)
			goto retry
		}

		err = metricsResponse.Unmarshal(body)
		if err != nil || len(metricsResponse.Metrics) == 0 {
			log.Println(err)
			log.Println("Error while parsing client's response")
			tries++
			time.Sleep(300 * time.Millisecond)
			goto retry
		}
	}

	return &metricsResponse, nil
}

func GetDetails(ips []string, cluster string, fetcher Fetcher) *pb.MetricDetailsResponse {

	response := &pb.MetricDetailsResponse{
		Metrics: make(map[string]*pb.MetricDetails),
	}

	for _, ip := range ips {
		url := "http://" + ip + "/metrics/details/?format=protobuf"
		fetcheddata, err := fetcher.FetchData(url)
		if err != nil {
			log.Println("timeout during fetching details")
			//return
		}
		if response == nil {
			continue
		}
		for m, v := range fetcheddata.Metrics {
			if r, ok := response.Metrics[m]; ok {

				if v.ModTime > r.ModTime {
					r.ModTime = v.ModTime
				}
				if v.Size_ > r.Size_ {
					r.Size_ = v.Size_
				}
			} else {
				response.Metrics[m] = v
			}
		}
	}
	return response
}

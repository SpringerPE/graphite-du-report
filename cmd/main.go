package main

import (
	"log"
	"fmt"
	"time"

	"io/ioutil"
	"net/http"
	_ "net/http/pprof"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

var errTimeout = fmt.Errorf("Max tries exceeded")

func fetchData(httpClient *http.Client, url string) (*pb.MetricDetailsResponse, error) {
	var metricsResponse pb.MetricDetailsResponse
	var response *http.Response
	var err error
	tries := 1

retry:
	if tries > 3 {
		log.Println("Tries exceeded while trying to fetch data")
		return nil, errTimeout
	}
	response, err = httpClient.Get(url)
	if err != nil {
		log.Println("Error during communication with client")
		tries++
		time.Sleep(300 * time.Millisecond)
		goto retry
	} else {
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println("Error while reading client's response")
			tries++
			time.Sleep(300 * time.Millisecond)
			goto retry
		}

		err = metricsResponse.Unmarshal(body)
		if err != nil || len(metricsResponse.Metrics) == 0 {
			log.Println("Error while parsing client's response")
			tries++
			time.Sleep(300 * time.Millisecond)
			goto retry
		}
	}

	return &metricsResponse, nil
}



func getDetails(ips []string, cluster string) *pb.MetricDetailsResponse {
	_ = &http.Client{Timeout: 120 * time.Second}
	response := &pb.MetricDetailsResponse{
		Metrics: make(map[string]*pb.MetricDetails),
	}
/*	responses := make([]*pb.MetricDetailsResponse, len(ips))
	fetchingLimiter := newLimiter(config.FetchPerCluster)

	var wg sync.WaitGroup
	for idx, ip := range ips {
		wg.Add(1)
		go func(i int, ip string) {
			fetchingLimiter.enter()
			defer fetchingLimiter.leave()
			defer wg.Done()
			url := "http://127.0.0.1:8080/metrics/details/?format=protobuf"
			data, err := fetchData(httpClient, url)
			if err != nil {
				logger.Error("timeout during fetching details")
				return
			}
			responses[i] = data
		}(idx, ip)
	}
	wg.Wait()

	maxCount := uint64(1)
	metricsReplicationCounter := make(map[string]uint64)
	for idx := range responses {
		if responses[idx] == nil {
			continue
		}

		response.FreeSpace += responses[idx].FreeSpace
		response.TotalSpace += responses[idx].TotalSpace

		for m, v := range responses[idx].Metrics {
			if r, ok := response.Metrics[m]; ok {
				metricsReplicationCounter[m]++
				if metricsReplicationCounter[m] > maxCount {
					maxCount = metricsReplicationCounter[m]
				}
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

	response.FreeSpace /= uint64(maxCount)
	response.TotalSpace /= uint64(maxCount)
*/
	return response
}

func main() {
	getDetails([]string{}, "")
}
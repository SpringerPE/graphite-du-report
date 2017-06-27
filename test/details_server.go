package main

import (
	"fmt"
	"github.com/ecooper/combinatoric"
	"net/http"
	"strings"
	"time"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"math/rand"
)

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func i2s(values []interface{}) string {
	s := ""
	for _, v := range values {
		s += fmt.Sprintf("%s", v)
	}
	return s
}

func generateMetric(iter *combinatoric.CombinationIterator, details *pb.MetricDetailsResponse, maxDepth int, remaining int) int {
	depth := random(1, maxDepth)
	baseName := i2s(iter.Next())
	for index := 0; index < depth; index++ {
		next := i2s(iter.Next())
		baseName = strings.Join([]string{baseName, next}, ".")
	}

	leaves := remaining
	if remaining > 1024 {
		leaves = random(1, 1024)
	}

	for index := 0; index < leaves; index++ {
		next := i2s(iter.Next())
		element := strings.Join([]string{baseName, next}, ".")
		details.Metrics[element] = &pb.MetricDetails{Size_: int64(1)}
	}
	return leaves
}

func detailsHandler(wr http.ResponseWriter, req *http.Request, response *pb.MetricDetailsResponse) {
	// URL: /metrics/details/?format=json
	req.ParseForm()
	format := req.FormValue("format")

	if format != "protobuf" && format != "protobuf3" {
		http.Error(wr, "Bad request (unsupported format)",
			http.StatusBadRequest)
		return
	}

	var err error
	var b []byte

	switch format {
	case "protobuf", "protobuf3":
		b, err = response.Marshal()
	}

	if err != nil {
		http.Error(wr, fmt.Sprintf("An internal error has occured: %s", err), http.StatusInternalServerError)
		return
	}
	wr.Write(b)

	return
}

func main() {
	rand.Seed(time.Now().Unix())

	elements := []interface{}{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

	// Create a new CombinationIterator of 2 elements using src
	iter, _ := combinatoric.Combinations(elements, 10)

	// Print the length of the iterator
	fmt.Printf("Expecting %d combinations:\n", iter.Len())

	details := &pb.MetricDetailsResponse{
		Metrics:    make(map[string]*pb.MetricDetails),
		FreeSpace:  uint64(1),
		TotalSpace: uint64(1),
	}

	remaining := 2000000

	for remaining > 0 {
		n := generateMetric(iter, details, 10, remaining)
		remaining -= n
	}

	http.HandleFunc("/metrics/details/", func(w http.ResponseWriter, r *http.Request) {
		detailsHandler(w, r, details)
	})

	http.ListenAndServe("127.0.0.1:8080", nil)
}

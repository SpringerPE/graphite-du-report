package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/SpringerPE/graphite-du-report/config"
	"github.com/SpringerPE/graphite-du-report/helper"
	"github.com/SpringerPE/graphite-du-report/reporter"

	"github.com/uber/go-torch/renderer"
)

type Worker struct {
	reader *reporter.TreeReader
	config *config.WorkerConfig
}

func NewWorker(reader *reporter.TreeReader, config *config.WorkerConfig) (*Worker, error) {
	worker := &Worker{reader: reader, config: config}
	return worker, nil
}

func (worker *Worker) HandleNodeSize(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	size, err := worker.reader.GetNodeSize(path)
	if err != nil {
		helper.ErrorResponse(w, "failed getting the node size", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("%d", size))
}

func (worker *Worker) HandleFlame(w http.ResponseWriter, r *http.Request) {
	flame, err := worker.reader.ReadFlameMap()
	if err != nil {
		helper.ErrorResponse(w, "failed reading the root node", err)
		return
	}
	flameInput := strings.Join(flame, "\n")
	//fmt.Println(flameInput)
	flameGraph, err := renderer.GenerateFlameGraph([]byte(flameInput), "--hash", "--countname=bytes")
	if err != nil {
		helper.ErrorResponse(w, "could not generate flame graph: %v", err)
		return
	}
	//TODO: set caching etag
	/**
		key := "somekey"
		e := `"` + key + `"`
		w.Header().Set("Etag", e)
	    	if match := r.Header.Get("If-None-Match"); match != "" {
	        		if strings.Contains(match, e) {
	            	w.WriteHeader(http.StatusNotModified)
	            	return
	        		}
	    	}
	**/
	w.Header().Set("Cache-Control", "max-age=300")
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write(flameGraph)
}

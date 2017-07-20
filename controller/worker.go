package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SpringerPE/graphite-du-report/caching"
	"github.com/SpringerPE/graphite-du-report/config"
	"github.com/SpringerPE/graphite-du-report/helper"
	"github.com/SpringerPE/graphite-du-report/reporter"

	"github.com/uber/go-torch/renderer"
)

type Worker struct {
	config *config.WorkerConfig
}

//TODO: make this a proper factory class
func (worker *Worker) createTreeReader() *reporter.TreeReader {
	config := worker.config

	reader := caching.NewRedisCaching(config.RedisAddr, config.RedisPasswd)
	treeReader, _ := reporter.NewTreeReader(config.RootName, reader)

	return treeReader
}

func NewWorker(config *config.WorkerConfig) (*Worker, error) {
	worker := &Worker{config: config}
	return worker, nil
}

func (worker *Worker) HandleNodeSize(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	reader := worker.createTreeReader()

	size, err := reader.GetNodeSize(path)
	if err != nil {
		helper.ErrorResponse(w, "failed getting the node size", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("%d", size))
}

func (worker *Worker) HandleFlame(w http.ResponseWriter, r *http.Request) {
	reader := worker.createTreeReader()

	flame, err := reader.ReadFlameMap()
	if err != nil {
		helper.ErrorResponse(w, "failed reading the root node", err)
		return
	}
	flameInput := strings.Join(flame, "\n")
	flameGraph, err := renderer.GenerateFlameGraph([]byte(flameInput), "--hash", "--countname=bytes")
	if err != nil {
		helper.ErrorResponse(w, "could not generate flame graph: %v", err)
		return
	}

	cacheSince := time.Now().Format(http.TimeFormat)
	cacheUntil := time.Now().Add(300 * time.Second).Format(http.TimeFormat)
	w.Header().Set("Last-Modified", cacheSince)
	w.Header().Set("Expires", cacheUntil)
	w.Header().Set("Cache-Control", "max-age=300, public")
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write(flameGraph)
}

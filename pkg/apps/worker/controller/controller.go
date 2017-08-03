package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/SpringerPE/graphite-du-report/pkg/caching"
	"github.com/SpringerPE/graphite-du-report/pkg/helper"

	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/reporter"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/config"
)

var templates = template.Must(template.ParseGlob("assets/worker/static/templates/*"))

type Worker struct {
	config *config.WorkerConfig
}

//TODO: make this a proper factory class
func (worker *Worker) createTreeReader() *reporter.TreeReader {
	conf := worker.config

	reader := caching.NewRedisCaching(conf.RedisAddr, conf.RedisPasswd, conf.RetrieveChildren)
	treeReader, _ := reporter.NewTreeReader(conf.RootName, reader)

	return treeReader
}

func NewWorker(config *config.WorkerConfig) (*Worker, error) {
	worker := &Worker{config: config}
	return worker, nil
}

func (worker *Worker) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_ = templates.ExecuteTemplate(w, "worker_home.html", make(map[interface{}]interface{}))
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
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	params := make(map[string]interface{})
	params["svg"] = ""

	_ = templates.ExecuteTemplate(w, "flame.html", params)
}

func (worker *Worker) HandleFoldedData(w http.ResponseWriter, r *http.Request) {
	reader := worker.createTreeReader()

	flame, err := reader.ReadFlameMap()
	if err != nil {
		helper.ErrorResponse(w, "failed reading the root node", err)
		return
	}

	flameInput := strings.Join(flame, "\n")

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(flameInput))
}

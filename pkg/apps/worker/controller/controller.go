package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/SpringerPE/graphite-du-report/pkg/helper"

	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/reporter"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/config"
)

type Worker struct {
	config *config.WorkerConfig
	treeReaderFactory reporter.TreeReaderFactory
	templates *template.Template
}

func NewWorker(config *config.WorkerConfig, treeReaderFactory reporter.TreeReaderFactory) (*Worker, error) {
	worker := &Worker{config: config, treeReaderFactory: treeReaderFactory}
	worker.templates = template.Must(template.ParseGlob(config.TemplatesFolder))
	return worker, nil
}

func (worker *Worker) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_ = worker.templates.ExecuteTemplate(w, "worker_home.html", make(map[interface{}]interface{}))
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

func (worker *Worker) HandleJsonData(w http.ResponseWriter, r *http.Request) {
	reader := worker.createTreeReader()
	jsonTree, err := reader.ReadJsonTree()
	if err != nil {
		helper.ErrorResponse(w, "failed reading the json for the tree", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(jsonTree))
}

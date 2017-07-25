package controller

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/SpringerPE/graphite-du-report/caching"
	"github.com/SpringerPE/graphite-du-report/config"
	"github.com/SpringerPE/graphite-du-report/helper"
	"github.com/SpringerPE/graphite-du-report/reporter"

	"github.com/uber/go-torch/renderer"
)

var flameTemplateString = `
<object class="svg" type="image/svg+xml" data="data:image/svg+xml;base64,{{.svgImage}}"/>
</object>
 `
var templates = template.Must(template.ParseGlob("static/templates/*"))

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
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	params := make(map[string]interface{})
	params["svg"] = ""

	_ = templates.ExecuteTemplate(w, "flame.html", params)
}

func (worker *Worker) HandleFlameImage(w http.ResponseWriter, r *http.Request) {
	t := template.New("Flame Image")
	tmpl, err := t.Parse(flameTemplateString)
	if err != nil {
		helper.ErrorResponse(w, "failed parsing image template", err)
		return
	}

	reader := worker.createTreeReader()

	flame, err := reader.ReadFlameMap()
	if err != nil {
		helper.ErrorResponse(w, "failed reading the root node", err)
		return
	}

	flameInput := strings.Join(flame, "\n")
	flameGraph, err := renderer.GenerateFlameGraph([]byte(flameInput),
		"--hash", "--countname=bytes")
	if err != nil {
		helper.ErrorResponse(w, "could not generate flame graph: %v", err)
		return
	}

	cacheSince := time.Now().Format(http.TimeFormat)
	cacheUntil := time.Now().Add(300 * time.Second).Format(http.TimeFormat)

	sEnc := base64.StdEncoding.EncodeToString(flameGraph)
	objects := make(map[string]interface{})
	objects["svgImage"] = sEnc

	w.Header().Set("Last-Modified", cacheSince)
	w.Header().Set("Expires", cacheUntil)
	w.Header().Set("Cache-Control", "max-age=300, public")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err = tmpl.Execute(w, objects)
}

package controller

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/SpringerPE/graphite-du-report/helper"
	"github.com/SpringerPE/graphite-du-report/renderer/config"

	torch "github.com/uber/go-torch/renderer"
)

var flameTemplateString = `
<object class="svg" type="image/svg+xml" data="data:image/svg+xml;base64,{{.svgImage}}"/>
</object>
 `

type Renderer struct {
	config *config.RendererConfig
}

func NewRenderer(config *config.RendererConfig) (*Renderer, error) {
	renderer := &Renderer{config: config}
	return renderer, nil
}

func (renderer *Renderer) HandleFlameImage(w http.ResponseWriter, r *http.Request) {
	t := template.New("Flame Image")
	tmpl, err := t.Parse(flameTemplateString)
	if err != nil {
		helper.ErrorResponse(w, "failed parsing image template", err)
		return
	}

	resp, err := http.Get(fmt.Sprintf("http://%s/folded", r.Host))
	if err != nil {
		helper.ErrorResponse(w, "failed getting the folded data representation", err)
		return
	}
	defer resp.Body.Close()
	flameInput, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		helper.ErrorResponse(w, "failed reading the body", err)
		return
	}

	flameGraph, err := torch.GenerateFlameGraph([]byte(flameInput),"--hash", "--countname=bytes")
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

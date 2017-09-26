package controller

import (
	"html/template"
	"net/http"

	"github.com/SpringerPE/graphite-du-report/pkg/apps/visualiser/config"
)

type Visualiser struct {
	config    *config.VisualiserConfig
	templates *template.Template
}

func NewVisualiser(config *config.VisualiserConfig) (*Visualiser, error) {
	visualiser := &Visualiser{config: config}
	visualiser.templates = template.Must(template.ParseGlob(config.TemplatesFolder))
	return visualiser, nil
}

func (visualiser *Visualiser) HandleFlame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	params := make(map[string]interface{})
	params["svg"] = ""

	_ = visualiser.templates.ExecuteTemplate(w, "flame.html", params)
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"net/http"
	"net/http/pprof"

	"github.com/SpringerPE/graphite-du-report/pkg/logging"

	"github.com/SpringerPE/graphite-du-report/pkg/apps/visualiser/config"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/visualiser/controller"

	"github.com/gorilla/mux"
	"gopkg.in/alecthomas/kingpin.v2"

)

var (
	profiling   = kingpin.Flag("profiling", "enable profiling via pprof").Default("false").OverrideDefaultFromEnvar("ENABLE_PPROF").Bool()
	bindAddress = kingpin.Flag("bind-address", "bind address for this server").Default("0.0.0.0").OverrideDefaultFromEnvar("BIND_ADDRESS").String()
	bindPort    = kingpin.Flag("bind-port", "bind port for this server").Default("6063").OverrideDefaultFromEnvar("PORT").String()
	basePath = kingpin.Flag("base-path", "base path for this component").Default("visualiser").OverrideDefaultFromEnvar("BASE_PATH").String()
	rendererPath = kingpin.Flag("renderer-path", "base path for the renderer component").Default("renderer").OverrideDefaultFromEnvar("RENDERER_PATH").String()
)

func attachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)
}

func attachStatic(router *mux.Router) {
	router.PathPrefix("/visualiser/static/").Handler(http.StripPrefix("/visualiser/static", http.FileServer(http.Dir("./assets/visualiser/static"))))
}

func main() {
	runVisualiser()
}

func runVisualiser() {
	kingpin.Parse()

	baseFolder, _ := os.Getwd()
	templateFolder := filepath.Join(baseFolder, "assets/visualiser/static/templates/*")

	visualiserConfig := &config.VisualiserConfig{
		BindPort:        *bindPort,
		TemplatesFolder: templateFolder,
		BasePath: *basePath,
		RendererPath: *rendererPath,
	}

	visualiser, _ := controller.NewVisualiser(visualiserConfig)

	router := mux.NewRouter()
	if *profiling {
		attachProfiler(router)
	}

	attachStatic(router)

	router.HandleFunc(filepath.Join("/", visualiserConfig.BasePath, "flame"), visualiser.HandleFlame).Methods("GET").Name("Flame")

	srv := &http.Server{
		Handler: router,
		Addr:    visualiserConfig.BindAddress + ":" + visualiserConfig.BindPort,
	}
	logging.LogStd(fmt.Sprintf("%s", srv.ListenAndServe()))
}

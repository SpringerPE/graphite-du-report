package main

import (
	"fmt"
	"path/filepath"
	"net/http"
	"net/http/pprof"

	"github.com/SpringerPE/graphite-du-report/pkg/logging"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/renderer/config"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/renderer/controller"

	"github.com/gorilla/mux"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	profiling   = kingpin.Flag("profiling", "enable profiling via pprof").Default("false").OverrideDefaultFromEnvar("ENABLE_PPROF").Bool()
	bindAddress = kingpin.Flag("bind-address", "bind address for this server").Default("0.0.0.0").OverrideDefaultFromEnvar("BIND_ADDRESS").String()
	bindPort    = kingpin.Flag("bind-port", "bind port for this server").Default("6062").OverrideDefaultFromEnvar("PORT").String()
	basePath = kingpin.Flag("base-path", "base path for this component").Default("renderer").OverrideDefaultFromEnvar("BASE_PATH").String()
)

func attachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)
}

func attachStatic(router *mux.Router) {
	router.PathPrefix("/renderer/static/").Handler(http.StripPrefix("/renderer/static/", http.FileServer(http.Dir("./renderer/static"))))
}

func main() {
	runRenderer()
}

func runRenderer() {
	kingpin.Parse()

	rendererConfig := &config.RendererConfig{
		BindAddress: *bindAddress,
		BindPort: *bindPort,
		BasePath: *basePath,
	}

	renderer, _ := controller.NewRenderer(rendererConfig)

	router := mux.NewRouter()
	if *profiling {
		attachProfiler(router)
	}

	router.HandleFunc(filepath.Join("/", rendererConfig.BasePath, "flame"), renderer.HandleFlameImage).Methods("GET").Name("Flame")

	srv := &http.Server{
		Handler: router,
		Addr:    rendererConfig.BindAddress + ":" + rendererConfig.BindPort,
	}
	logging.LogStd(fmt.Sprintf("%s", srv.ListenAndServe()))
}

package main

import (
	"fmt"
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
)

func attachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)
}

func attachStatic(router *mux.Router) {
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
}

func main() {
	runRenderer()
}

func runRenderer() {
	kingpin.Parse()

	config := &config.RendererConfig{
		BindPort: *bindPort,
	}

	renderer, _ := controller.NewRenderer(config)

	router := mux.NewRouter()
	if *profiling {
		attachProfiler(router)
	}

	router.HandleFunc("/render/flame", renderer.HandleFlameImage).Methods("GET").Name("Flame")

	srv := &http.Server{
		Handler: router,
		Addr:    config.BindAddress + ":" + config.BindPort,
	}
	logging.LogStd(fmt.Sprintf("%s", srv.ListenAndServe()))
}

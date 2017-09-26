package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/SpringerPE/graphite-du-report/pkg/logging"

	"github.com/SpringerPE/graphite-du-report/pkg/apps/visualiser/config"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/visualiser/controller"

	"github.com/gorilla/mux"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
)

var (
	profiling   = kingpin.Flag("profiling", "enable profiling via pprof").Default("false").OverrideDefaultFromEnvar("ENABLE_PPROF").Bool()
	bindAddress = kingpin.Flag("bind-address", "bind address for this server").Default("0.0.0.0").OverrideDefaultFromEnvar("BIND_ADDRESS").String()
	bindPort    = kingpin.Flag("bind-port", "bind port for this server").Default("6063").OverrideDefaultFromEnvar("PORT").String()
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
	}

	visualiser, _ := controller.NewVisualiser(visualiserConfig)

	router := mux.NewRouter()
	if *profiling {
		attachProfiler(router)
	}

	attachStatic(router)

	router.HandleFunc("/visualiser/flame", visualiser.HandleFlame).Methods("GET").Name("Flame")

	srv := &http.Server{
		Handler: router,
		Addr:    visualiserConfig.BindAddress + ":" + visualiserConfig.BindPort,
	}
	logging.LogStd(fmt.Sprintf("%s", srv.ListenAndServe()))
}

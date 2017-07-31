package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/SpringerPE/graphite-du-report/caching"
	"github.com/SpringerPE/graphite-du-report/helper"
	"github.com/SpringerPE/graphite-du-report/logging"

	"github.com/SpringerPE/graphite-du-report/updater/reporter"
	"github.com/SpringerPE/graphite-du-report/updater/config"
)

//TODO: make this a proper factory class
func (up *Updater) createBuilderTree() *reporter.Tree {
	config := up.config

	builder := caching.NewMemBuilder()

	updater := caching.NewRedisCaching(config.RedisAddr, config.RedisPasswd)
	updater.SetNumBulkScans(config.BulkScans)
	locker := updater

	tree, _ := reporter.NewTree(config.RootName, builder, updater, locker)
	tree.SetNumUpdateRoutines(config.UpdateRoutines)
	tree.SetNumBulkUpdates(config.BulkUpdates)

	return tree
}

type Updater struct {
	config *config.UpdaterConfig
}

func NewUpdater(config *config.UpdaterConfig) *Updater {
	return &Updater{
		config: config,
	}
}

func (up *Updater) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	return
}

func (up *Updater) PopulateDetails(w http.ResponseWriter, r *http.Request) {
	config := up.config
	tree := up.createBuilderTree()
	fetcher := reporter.NewDataFetcher(120*time.Second, 3)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//generate a secret for the update lock
	lockName := "update_lock"
	secret, err := helper.GenerateSecret()
	if err != nil {
		helper.ErrorResponse(w, "error while generating lock secret", err)
		return
	}
	// try to acquire the lock
	ok, err := tree.WriteLock(lockName, secret, 120)
	if !ok {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Another update operation is currently running")
		return
	}
	defer func() {
		ok, err = tree.ReleaseLock(lockName, secret)
		if !ok {
			logging.LogError("failed releasing the update lock", err)
		}
	}()

	response, err := reporter.GetDetails(config.Servers, "", fetcher)
	if err != nil {
		helper.ErrorResponse(w, "error while contacting carbonserver", err)
		return
	}
	logging.LogStd(fmt.Sprintf("%s", "Tree building started"))
	// Construct the tree from the metrics response first
	err = tree.ConstructTree(response)
	if err != nil {
		helper.ErrorResponse(w, "cannot construct the tree from the metrics response", err)
		return
	}
	logging.LogStd(fmt.Sprintf("%s", "Tree building finished"))
	err = tree.Persist()
	if err != nil {
		helper.ErrorResponse(w, "tree populate failed", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	return
}

func (up *Updater) Cleanup(w http.ResponseWriter, r *http.Request) {
	config := up.config
	tree := up.createBuilderTree()

	err := tree.Cleanup(config.RootName)
	if err != nil {
		helper.ErrorResponse(w, "failed cleaning up", err)
		return
	}
	logging.LogStd(fmt.Sprintf("%s", "cleanup finished"))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("%s", "OK"))
}

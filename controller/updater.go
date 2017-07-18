package controller

import (
	"fmt"
	"net/http"

	"github.com/SpringerPE/graphite-du-report/config"
	"github.com/SpringerPE/graphite-du-report/helper"
	"github.com/SpringerPE/graphite-du-report/logging"
	"github.com/SpringerPE/graphite-du-report/reporter"
)

type Updater struct {
	tree    *reporter.Tree
	fetcher reporter.Fetcher
	config  *config.UpdaterConfig
}

func NewUpdater(tree *reporter.Tree, fetcher reporter.Fetcher, config *config.UpdaterConfig) (*Updater, error) {
	up := &Updater{
		tree:    tree,
		fetcher: fetcher,
		config:  config,
	}
	return up, nil
}

func (up *Updater) PopulateDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//generate a secret for the update lock
	lockName := "update_lock"
	secret, err := helper.GenerateSecret()
	if err != nil {
		helper.ErrorResponse(w, "error while generating lock secret", err)
		return
	}
	// try to acquire the lock
	ok, err := up.tree.WriteLock(lockName, secret, 120)
	if !ok {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Another update operation is currently running")
		return
	}
	defer func() {
		ok, err = up.tree.ReleaseLock(lockName, secret)
		if !ok {
			logging.LogError("failed releasing the update lock", err)
		}
	}()

	response := reporter.GetDetails(up.config.Servers, "", up.fetcher)
	logging.LogStd(fmt.Sprintf("%s", "Tree building started"))
	// Construct the tree from the metrics response first
	err = up.tree.ConstructTree(response)
	if err != nil {
		helper.ErrorResponse(w, "cannot construct the tree from the metrics response", err)
		return
	}
	logging.LogStd(fmt.Sprintf("%s", "Tree building finished"))
	err = up.tree.Persist()
	if err != nil {
		helper.ErrorResponse(w, "tree populate failed", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	return
}

func (up *Updater) Cleanup(w http.ResponseWriter, r *http.Request) {
	logging.LogStd(fmt.Sprintf("%s", "Tree cleanup started"))

	err := up.tree.Cleanup(up.config.RootName)
	if err != nil {
		helper.ErrorResponse(w, "failed cleaning up", err)
		return
	}
	logging.LogStd(fmt.Sprintf("%s", "cleanup finished"))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("%s", "OK"))
}
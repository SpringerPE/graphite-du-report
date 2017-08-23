package controller_test

import (
	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/controller"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/config"


	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"net/http/httptest"
)

var _ = Describe("Controller", func() {

	var (
		worker *controller.Worker
		handler http.HandlerFunc
		workerConfig *config.WorkerConfig
		req *http.Request
		rr *httptest.ResponseRecorder
		err error
	)

	BeforeEach(func(){
		workerConfig = config.DefaultWorkerConfig()
		workerConfig.TemplatesFolder = "../../../../assets/worker/static/templates/*"
		treeReaderFactory := MockTreeReaderFactory{}
		worker, _ = controller.NewWorker(workerConfig, treeReaderFactory)
	})

	JustBeforeEach(func(){
		rr = httptest.NewRecorder()
	})

	Context("when a handle node size request is received", func(){

		JustBeforeEach(func(){
			req, err = http.NewRequest("GET", "/get_size", nil)
			Expect(err).To(BeNil())
			handler = http.HandlerFunc(worker.HandleNodeSize)
		})

		It("it serves the size of the requested node", func(){
			q := req.URL.Query()
			q.Add("path", "root")
			req.URL.RawQuery = q.Encode()
			// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			handler.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			Expect(rr.Code).To(Equal(http.StatusOK))

			// Check the response body is what we expect.
			Expect(rr.Body.String()).To(Equal("10"))
		})

		It("fails with an error when the requested node does not exist", func(){
			q := req.URL.Query()
			q.Add("path", "not_existent")
			req.URL.RawQuery = q.Encode()

			handler.ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusInternalServerError))
		})
	})

})

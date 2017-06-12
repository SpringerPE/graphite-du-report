package reporter_test

import (
	//"fmt"
	"github.com/SpringerPE/graphite-du-report/reporter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

var _ = Describe("Reporter", func() {

	var (
		response *pb.MetricDetailsResponse
	)

	BeforeEach(func() {
		response = &pb.MetricDetailsResponse{
			map[string]*pb.MetricDetails{
				"team1.metric1":              &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
				"team2.metric1":              &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
				"team2.metric2":              &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
				"team1.stats.metric1":        &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
				"team2.stats.metric1":        &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
				"team1.stats.gauges.metric1": &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
				"team2.stats.gauges.metric1": &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
			},
			38527414272,
			42241163264,
		}
	})

	Describe("Can construct a tree starting from a MetricDetails response", func() {
		Context("If the response is well formed", func() {
			It("Can construct the tree", func() {
				root := &reporter.Node{Name: "root", Children: map[string]*reporter.Node{}}
				reporter.ConstructTree(root, response)
				rootChildren := root.Children
				Expect(rootChildren).To(HaveLen(2))
				for _, key := range []string{"team1", "team2"} {
					Expect(rootChildren).To(HaveKey(key))
					Expect(rootChildren[key].Leaf).To(BeFalse())
					Expect(rootChildren[key].Size).To(Equal(int64(0)))
				}
				team1Children := rootChildren["team1"].Children
				for _, key := range []string{"metric1"} {
					Expect(team1Children).To(HaveKey(key))
					Expect(team1Children[key].Leaf).To(BeTrue())
					Expect(team1Children[key].Size).To(Equal(int64(520192)))
				}
				for _, key := range []string{"stats"} {
					Expect(team1Children).To(HaveKey(key))
					Expect(team1Children[key].Leaf).To(BeFalse())
				}
			})
		})
	})

	Describe("Can update the nodes metadata in a tree during a visit", func() {
		Context("Given the root of a tree", func() {
			It("Can update the metadata", func() {
				root := &reporter.Node{Name: "root", Children: map[string]*reporter.Node{}}
				reporter.ConstructTree(root, response)
				reporter.UpdateSize(root)
				rootChildren := root.Children
				Expect(rootChildren["team1"].Size).To(Equal(int64(1560576)))
				Expect(rootChildren["team2"].Size).To(Equal(int64(2080768)))
				team1Children := rootChildren["team1"].Children
				Expect(team1Children["metric1"].Size).To(Equal(int64(520192)))
				Expect(team1Children["stats"].Size).To(Equal(int64(1040384)))
			})
		})
	})

	Describe("Can get the details", func() {

		var (
			response1, response2, response3 *pb.MetricDetailsResponse
		)

		BeforeEach(func() {
			response1 = &pb.MetricDetailsResponse{
				map[string]*pb.MetricDetails{
					"team1.metric1":              &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
					"team1.stats.metric1":        &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
					"team1.stats.gauges.metric1": &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
				},
				38527414272,
				42241163264,
			}
			response2 = &pb.MetricDetailsResponse{
				map[string]*pb.MetricDetails{
					"team2.metric1":              &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
					"team2.metric2":              &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
					"team2.stats.metric1":        &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
					"team2.stats.gauges.metric1": &pb.MetricDetails{Size_: 520192, ModTime: 1497262565, ATime: 1497262565},
				},
				38527414272,
				42241163264,
			}
			response3 = &pb.MetricDetailsResponse{
				map[string]*pb.MetricDetails{
					"team1.metric1":              &pb.MetricDetails{Size_: 520192, ModTime: 1497262566, ATime: 1497262566},
					"team1.stats.metric1":        &pb.MetricDetails{Size_: 520193, ModTime: 1497262566, ATime: 1497262566},
					"team1.stats.gauges.metric1": &pb.MetricDetails{Size_: 520192, ModTime: 1497262564, ATime: 1497262565},
				},
				38527414272,
				42241163264,
			}

		})

		Context("Given multiple ip addresses", func() {
			It("return a valid response object", func() {
				fetcher := NewFakeDataFetcher()
				fetcher.Responses["http://127.0.0.1:8080/metrics/details/?format=protobuf"] = response1
				fetcher.Responses["http://127.0.0.2:8080/metrics/details/?format=protobuf"] = response2

				response := reporter.GetDetails([]string{"127.0.0.1", "127.0.0.2"}, "", fetcher)
				metricsList := []string{
					"team1.metric1", "team1.stats.metric1", "team1.stats.gauges.metric1",
					"team2.metric1", "team2.metric2", "team2.stats.metric1", "team2.stats.gauges.metric1",
				}
				for _, metric := range metricsList {
					Expect(response.Metrics).To(HaveKey(metric))
				}

			})

			It("overwrites older metrics with most recents", func() {
				fetcher := NewFakeDataFetcher()
				fetcher.Responses["http://127.0.0.1:8080/metrics/details/?format=protobuf"] = response1
				fetcher.Responses["http://127.0.0.2:8080/metrics/details/?format=protobuf"] = response3

				response := reporter.GetDetails([]string{"127.0.0.1", "127.0.0.2"}, "", fetcher)
				Expect(response.Metrics).To(HaveLen(3))
				Expect(response.Metrics["team1.metric1"].ModTime).To(Equal(int64(1497262566)))
				Expect(response.Metrics["team1.stats.metric1"].Size_).To(Equal(int64(520193)))
				Expect(response.Metrics["team1.stats.gauges.metric1"].ModTime).To(Equal(int64(1497262565)))
			})
		})
	})
})

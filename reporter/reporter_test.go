package reporter_test

import (
	"github.com/SpringerPE/graphite-du-report/caching"
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

	Describe("Tree Builder", func() {

		var (
			builder                  caching.TreeBuilder
			updater                  caching.TreeUpdater
			reader                   caching.TreeReader
			locker                   caching.Locker
			tree                     *reporter.Tree
			readerTree               *reporter.TreeReader
			err, buildErr, readerErr error
		)

		BeforeEach(func() {

		})

		JustBeforeEach(func() {
			builder = caching.NewMemBuilder()
			updater = NewMockUpdater()
			reader = updater.(caching.TreeReader)
			locker = NewMockLocker()
			tree, err = reporter.NewTree("root", builder, updater, locker)
			readerTree, readerErr = reporter.NewTreeReader("root", reader)
			buildErr = tree.ConstructTree(response)
		})

		Context("given a MetricsDetails response", func() {

			It("should not error when creating the tree", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should not error when creating the tree reader", func() {
				Expect(readerErr).NotTo(HaveOccurred())
			})

			It("should not error when building the tree", func() {
				Expect(buildErr).NotTo(HaveOccurred())
			})

			It("should be able to construct the metrics tree", func() {

				root, _ := tree.GetNode(tree.RootName)

				childrenNames := []string{"team1", "team2"}

				//tree should have two children named team1 and team2
				Expect(root.Children).To(HaveLen(2))
				for _, key := range childrenNames {
					Expect(root.Children).To(ContainElement(key))
				}

				//tree should two nodes named stats
				for _, key := range childrenNames {
					child, err := tree.GetNodeFromRoot(key + ".stats")
					Expect(err).To(BeNil())
					Expect(child.Leaf).To(BeFalse())
				}

				//tree should not contain original leaves (metric files)
				for _, key := range childrenNames {
					_, err := tree.GetNodeFromRoot(key + ".metric1")
					Expect(err).NotTo(BeNil())
				}

				//nodes 1 level up the leaves should be marked as leaves
				for _, key := range childrenNames {
					child, err := tree.GetNodeFromRoot(key + ".stats.gauges")
					Expect(err).To(BeNil())
					Expect(child.Leaf).To(BeTrue())
				}
			})

			It("should persist the data via the TreeUpdater", func() {
				persistErr := tree.Persist()
				Expect(persistErr).To(BeNil())

				names := []string{
					"root",
					"root.team1",
					"root.team2",
					"root.team1.stats",
					"root.team2.stats",
					"root.team1.stats.gauges",
					"root.team2.stats.gauges",
				}
				for _, name := range names {
					node, readErr := readerTree.ReadNode(name)
					Expect(readErr).To(BeNil())
					Expect(node.Name).To(Equal(name))
				}
			})

			It("should populate the tree with the correct metadata", func() {

				team1, _ := tree.GetNodeFromRoot("team1")
				team2, _ := tree.GetNodeFromRoot("team2")

				Expect(team1.Size).To(Equal(int64(1560576)))
				Expect(team2.Size).To(Equal(int64(2080768)))

				stats, _ := tree.GetNodeFromRoot("team2.stats")
				Expect(stats.Size).To(Equal(int64(1040384)))
			})
		})
	})

	Describe("Fetcher", func() {
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

		Context("given multiple ip addresses", func() {
			It("should return a valid response object", func() {
				fetcher := NewFakeDataFetcher()
				fetcher.Responses["http://127.0.0.1:8080/metrics/details/?format=protobuf3"] = response1
				fetcher.Responses["http://127.0.0.2:8080/metrics/details/?format=protobuf3"] = response2

				response := reporter.GetDetails([]string{"127.0.0.1:8080", "127.0.0.2:8080"}, "", fetcher)
				metricsList := []string{
					"team1.metric1", "team1.stats.metric1", "team1.stats.gauges.metric1",
					"team2.metric1", "team2.metric2", "team2.stats.metric1", "team2.stats.gauges.metric1",
				}
				for _, metric := range metricsList {
					Expect(response.Metrics).To(HaveKey(metric))
				}

			})

			It("should overwrite older metrics with most recents", func() {
				fetcher := NewFakeDataFetcher()
				fetcher.Responses["http://127.0.0.1:8080/metrics/details/?format=protobuf3"] = response1
				fetcher.Responses["http://127.0.0.2:8080/metrics/details/?format=protobuf3"] = response3

				response := reporter.GetDetails([]string{"127.0.0.1:8080", "127.0.0.2:8080"}, "", fetcher)
				Expect(response.Metrics).To(HaveLen(3))
				Expect(response.Metrics["team1.metric1"].ModTime).To(Equal(int64(1497262566)))
				Expect(response.Metrics["team1.stats.metric1"].Size_).To(Equal(int64(520193)))
				Expect(response.Metrics["team1.stats.gauges.metric1"].ModTime).To(Equal(int64(1497262565)))
			})
		})
	})
})

package caching_test

import (
	. "github.com/SpringerPE/graphite-du-report/pkg/caching"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rafaeljusto/redigomock"
)

var _ = Describe("Builder", func() {

	var (
		builder  TreeBuilder
		nodes    map[string]*Node
		notAdded *Node
	)

	BeforeEach(func() {
		nodes = make(map[string]*Node)
		nodes["root"] = &Node{
			Name:     "root",
			Leaf:     false,
			Size:     int64(1),
			Children: []string{},
		}
		nodes["root.team1"] = &Node{
			Name:     "root.team1",
			Leaf:     true,
			Size:     int64(1),
			Children: []string{},
		}
		nodes["root.team2"] = &Node{
			Name:     "root.team2",
			Leaf:     true,
			Size:     int64(1),
			Children: []string{},
		}

		notAdded = &Node{
			Name:     "not_added",
			Leaf:     true,
			Size:     int64(1),
			Children: []string{},
		}
	})

	Context("given a list of nodes and a mem builder", func() {

		JustBeforeEach(func() {
			builder = NewMemBuilder()
		})

		It("should generate an error if the node does not exist", func() {
			node, err := builder.GetNode("not_existent_node")
			Expect(err).To(HaveOccurred())
			Expect(node).To(BeNil())
		})

		It("should generate an error if we try to add a child to a not existent node", func() {
			err := builder.AddChild(notAdded, "child")
			Expect(err).To(HaveOccurred())
		})

		It("should be possible to add them", func() {
			builder.AddNode(nodes["root"])
			builder.AddChild(nodes["root"], "team1")
			builder.AddChild(nodes["root"], "team2")

			builder.AddNode(nodes["root.team1"])
			builder.AddNode(nodes["root.team2"])

			root, err := builder.GetNode("root")
			Expect(err).NotTo(HaveOccurred())
			Expect(root.Children).To(HaveLen(2))
			Expect(root.Children).To(ContainElement("team1"))
			Expect(root.Children).To(ContainElement("team2"))
		})
	})
})

var _ = Describe("Updater", func() {

	var (
		updater TreeUpdater
	)
	Context("given a list of nodes and a redis reader", func() {

		var (
			mockRedisConn *redigomock.Conn
			mockRedisPool Pool
		)

		JustBeforeEach(func() {
			mockRedisConn = redigomock.NewConn()

			mockRedisPool = newMockPool(mockRedisConn)

			updater = &RedisCaching{
				Pool:      mockRedisPool,
				BulkScans: 10,
			}
		})

		It("should return the correct version", func() {
			_ = mockRedisConn.Command("GET", "version").Expect("1")
			version, _ := updater.Version()

			Expect(version).To(Equal("1"))
		})

		It("should delete only the previous version keys", func() {
			_ = mockRedisConn.Command("GET", "version").Expect("1")
			mockRedisConn.Command("SCAN", int64(0), "count", 5000).Expect(
				[]interface{}{
					int64(10),
					[]interface{}{
						[]byte("version"),
						[]byte("0:root"),
						[]byte("1:root"),
					},
				})
			mockRedisConn.Command("SCAN", int64(10), "count", 5000).Expect([]interface{}{
				int64(0),
				[]interface{}{
					[]byte("anotherKey"),
					[]byte("0:folded"),
					[]byte("1:folded"),
				},
			})
			mockRedisConn.GenericCommand("MULTI").Expect("ok")
			mockRedisConn.GenericCommand("EXEC").Expect("ok")
			delRootCMD := mockRedisConn.Command("DEL", "0:root").Expect("ok")
			delFoldedCMD := mockRedisConn.Command("DEL", "0:folded").Expect("ok")

			err := updater.Cleanup("root")
			Expect(err).To(BeNil())
			Expect(delFoldedCMD.Called).To(BeTrue())
			Expect(delRootCMD.Called).To(BeTrue())
		})
	})
})

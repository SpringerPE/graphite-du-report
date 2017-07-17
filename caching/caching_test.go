package caching_test

import (
	. "github.com/SpringerPE/graphite-du-report/caching"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Caching", func() {

	var (
		memBuilder TreeBuilder
		nodes      map[string]*Node
		notAdded   *Node
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

	JustBeforeEach(func() {
		memBuilder = NewMemBuilder()
	})

	Context("given a list of nodes", func() {

		It("should generate an error if the node does not exist", func() {
			node, err := memBuilder.GetNode("not_existent_node")
			Expect(err).To(HaveOccurred())
			Expect(node).To(BeNil())
		})

		It("should generate an error if we try to add a child to a not existent node", func() {
			err := memBuilder.AddChild(notAdded, "child")
			Expect(err).To(HaveOccurred())
		})

		It("should be possible to add them", func() {
			memBuilder.AddNode(nodes["root"])
			memBuilder.AddChild(nodes["root"], "team1")
			memBuilder.AddChild(nodes["root"], "team2")

			memBuilder.AddNode(nodes["root.team1"])
			memBuilder.AddNode(nodes["root.team2"])

			root, err := memBuilder.GetNode("root")
			Expect(err).NotTo(HaveOccurred())
			Expect(root.Children).To(HaveLen(2))
			Expect(root.Children).To(ContainElement("team1"))
			Expect(root.Children).To(ContainElement("team2"))
		})
	})
})

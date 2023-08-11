package gwf

import (
	"fmt"
	"testing"
)

func TestTrie(t *testing.T) {
	root := &treeNode{name: "/", children: make([]*treeNode, 0)}
	root.Put("/user/get/:id")
	root.Put("/user/export")
	root.Put("/user/export/hello")
	root.Put("/order/export/aaa")

	node := root.Get("/user/get/1")
	fmt.Println(node)
	node = root.Get("/user/export/1")
	fmt.Println(node)
	node = root.Get("/user/export/hello")
	fmt.Println(node)
	node = root.Get("/user/export")
	fmt.Println(node)
	node = root.Get("/order/export")
	fmt.Println(node)
}

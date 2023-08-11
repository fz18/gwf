package gwf

import "strings"

type treeNode struct {
	name     string
	children []*treeNode
	path     string
	isEnd    bool
}

func (t *treeNode) Put(path string) {
	strs := strings.Split(path, "/")
	// delete first space
	strs = strs[1:]

	node := t
	for _, str := range strs {
		isMatch := false
		for _, child := range node.children {
			if child.name == str {
				node = child
				isMatch = true
				break
			}
		}
		if !isMatch {
			newNode := treeNode{name: str}
			node.children = append(node.children, &newNode)
			node = &newNode
		}
	}
	node.isEnd = true
}

func (t *treeNode) Get(path string) *treeNode {
	strs := strings.Split(path, "/")
	strs = strs[1:]

	node := t
	path = ""
	for i, str := range strs {
		for _, child := range node.children {
			if child.name == str ||
				child.name == "*" ||
				strings.Contains(child.name, ":") {
				node = child
				path += "/" + child.name
				node.path = path
				if i == len(strs)-1 {
					return node
				}
				break
			}
		}
	}
	return nil
}

package daox

import (
	"context"
	"strings"
)

// RelFiller 关联数据填充器
type RelFiller[T Model] func(ctx context.Context, src []T, n *PreloadNode) error

// PreloadNode 预加载节点
type PreloadNode struct {
	Node map[string]*PreloadNode
}

func parsePreloadPath(paths ...string) *PreloadNode {
	if len(paths) == 0 {
		return nil
	}
	root := &PreloadNode{
		Node: map[string]*PreloadNode{},
	}
	for _, path := range paths {
		parts := strings.Split(path, ".")
		node := root
		for _, part := range parts {
			if node.Node[part] == nil {
				node.Node[part] = &PreloadNode{
					Node: map[string]*PreloadNode{},
				}
			}
			node = node.Node[part]
		}
	}
	return root
}

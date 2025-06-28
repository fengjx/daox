package daox

import (
	"testing"
)

func TestParsePreloadPath(t *testing.T) {
	tests := []struct {
		name   string
		paths  []string
		assert func(root *PreloadNode, t *testing.T)
	}{
		{
			name:  "single level",
			paths: []string{"Orders"},
			assert: func(root *PreloadNode, t *testing.T) {
				if root.Node["Orders"] == nil {
					t.Errorf("Orders should be IsLeaf")
				}
			},
		},
		{
			name:  "multi level",
			paths: []string{"Orders.Items.Goods"},
			assert: func(root *PreloadNode, t *testing.T) {
				o := root.Node["Orders"]
				if o == nil {
					t.Errorf("Orders should exist and not be IsLeaf")
				}
				i := o.Node["Items"]
				if i == nil {
					t.Errorf("Items should exist and not be IsLeaf")
				}
				g := i.Node["Goods"]
				if g == nil {
					t.Errorf("Goods should exist and be IsLeaf")
				}
			},
		},
		{
			name:  "branch",
			paths: []string{"Orders.Items", "Orders.Payments"},
			assert: func(root *PreloadNode, t *testing.T) {
				o := root.Node["Orders"]
				if o == nil {
					t.Errorf("Orders should exist")
				}
				if o.Node["Items"] == nil {
					t.Errorf("Items should exist and be IsLeaf")
				}
				if o.Node["Payments"] == nil {
					t.Errorf("Payments should exist and be IsLeaf")
				}
			},
		},
		{
			name:  "merge and deep",
			paths: []string{"Orders.Items.Goods", "Orders.Items"},
			assert: func(root *PreloadNode, t *testing.T) {
				o := root.Node["Orders"]
				i := o.Node["Items"]
				g := i.Node["Goods"]
				if i == nil {
					t.Errorf("Items should exist and be IsLeaf")
				}
				if g == nil {
					t.Errorf("Goods should exist and be IsLeaf")
				}
			},
		},
		{
			name:  "multiple roots",
			paths: []string{"Orders", "Cards"},
			assert: func(root *PreloadNode, t *testing.T) {
				if root.Node["Orders"] == nil {
					t.Errorf("Orders should be IsLeaf")
				}
				if root.Node["Cards"] == nil {
					t.Errorf("Cards should be IsLeaf")
				}
			},
		},
		{
			name:  "empty",
			paths: []string{},
			assert: func(root *PreloadNode, t *testing.T) {
				if root != nil {
					t.Errorf("root should be nil for empty input")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := parsePreloadPath(tt.paths...)
			tt.assert(root, t)
		})
	}
}

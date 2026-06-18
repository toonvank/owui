package api

import "testing"

func TestFeaturesFromModel(t *testing.T) {
	m := Model{
		ID: "mark-with-a-k",
		Info: map[string]any{
			"meta": map[string]any{
				"capabilities": map[string]any{
					"web_search":       true,
					"code_interpreter": true,
					"memory":           false,
				},
			},
		},
	}
	features := FeaturesFromModel(m)
	if features["web_search"] != true {
		t.Fatalf("expected web_search")
	}
	if features["code_interpreter"] != true {
		t.Fatalf("expected code_interpreter")
	}
	if _, ok := features["memory"]; ok {
		t.Fatalf("memory should be omitted when false")
	}
}
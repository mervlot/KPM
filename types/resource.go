package types

type Resource struct {
	Group   *string `json:"group"`
	Name    string  `json:"name"`
	Version *string `json:"version"`
	Source  string  `json:"source"`
	Type    string  `json:"type"`
	Domain  *string `json:"domain"`
	Path    string  `json:"path"`
	LPath   string  `json:"lpath"`
	GPath   string  `json:"gpath"`
	URL     *string `json:"url"`
	Hash    string  `json:"hash"`
}

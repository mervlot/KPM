package types

type Resource struct {
	Group   *string `json:"group,omitempty"`
	Name    string  `json:"name"`
	Version *string `json:"version,omitempty"`
	Source  string  `json:"source"`
	Type    string  `json:"type"`
	Domain  *string `json:"domain,omitempty"`
	Path    *string `json:"path,omitempty"`
	LPath   string  `json:"lpath"`
	GPath   *string `json:"gpath,omitempty"`
	URL     *string `json:"url,omitempty"`
	Hash    string  `json:"hash"`
}

type ResourceFile struct {
	Version   string     `json:"kpm version"`
	Resources []Resource `json:"resources"`
}

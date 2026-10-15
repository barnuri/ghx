package release

import "strings"

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

func (rel Release) Version() string {
	return strings.TrimPrefix(rel.TagName, "v")
}

func (rel Release) Find(name string) (Asset, bool) {
	for _, asset := range rel.Assets {
		if asset.Name == name {
			return asset, true
		}
	}
	return Asset{}, false
}

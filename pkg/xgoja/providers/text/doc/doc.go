package doc

import "embed"

//go:embed *.md
var docs embed.FS

// FS returns the embedded Glazed help pages for the goja-text xgoja provider.
func FS() embed.FS {
	return docs
}

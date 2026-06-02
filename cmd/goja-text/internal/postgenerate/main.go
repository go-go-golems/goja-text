package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	cwd, err := os.Getwd()
	must(err)
	root, err := filepath.Abs(filepath.Join(cwd, "../.."))
	must(err)
	gomodPath := filepath.Join(cwd, "go.mod")
	data, err := os.ReadFile(gomodPath)
	must(err)

	content := string(data)
	content = strings.ReplaceAll(content,
		"replace github.com/go-go-golems/goja-text => "+filepath.ToSlash(root),
		"replace github.com/go-go-golems/goja-text => ../..",
	)
	content = strings.ReplaceAll(content,
		"replace github.com/go-go-golems/goja-text => "+root,
		"replace github.com/go-go-golems/goja-text => ../..",
	)
	if !strings.Contains(content, "tool github.com/go-go-golems/go-go-goja/cmd/xgoja") {
		content = strings.TrimRight(content, "\n") + "\n\ntool github.com/go-go-golems/go-go-goja/cmd/xgoja\n"
	}
	must(os.WriteFile(gomodPath, []byte(content), 0o644))
}

func must(err error) {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

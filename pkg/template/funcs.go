package template

import (
	"fmt"
	"sort"
	texttemplate "text/template"

	"github.com/Masterminds/sprig"
	glazedtemplating "github.com/go-go-golems/glazed/pkg/helpers/templating"
)

var defaultFuncSets = []string{"sprig", "glazed"}
var allowedFuncSets = map[string]struct{}{
	"none":   {},
	"sprig":  {},
	"glazed": {},
}

func normalizeFuncSets(names []string) []string {
	if len(names) == 0 {
		return append([]string(nil), defaultFuncSets...)
	}
	ret := make([]string, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		ret = append(ret, name)
	}
	if len(ret) == 0 {
		return append([]string(nil), defaultFuncSets...)
	}
	return ret
}

func validateFuncSets(names []string) []string {
	var errs []string
	if len(names) > 1 {
		for _, name := range names {
			if name == "none" {
				errs = append(errs, "func set \"none\" cannot be combined with other function sets")
				break
			}
		}
	}
	for _, name := range names {
		if _, ok := allowedFuncSets[name]; !ok {
			errs = append(errs, fmt.Sprintf("unknown func set %q", name))
		}
	}
	return errs
}

func funcMapFor(mode Mode, names []string) (texttemplate.FuncMap, error) {
	sets := normalizeFuncSets(names)
	if errs := validateFuncSets(sets); len(errs) > 0 {
		return nil, fmt.Errorf("template funcs: %s", joinErrors(errs))
	}
	out := texttemplate.FuncMap{}
	for _, name := range sets {
		switch name {
		case "none":
			return out, nil
		case "sprig":
			if mode == ModeHTML {
				mergeFuncMap(out, sprig.HtmlFuncMap())
			} else {
				mergeFuncMap(out, sprig.TxtFuncMap())
			}
		case "glazed":
			mergeFuncMap(out, glazedtemplating.TemplateFuncs)
		}
	}
	return out, nil
}

func mergeFuncMap(dst, src texttemplate.FuncMap) {
	keys := make([]string, 0, len(src))
	for key := range src {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		dst[key] = src[key]
	}
}

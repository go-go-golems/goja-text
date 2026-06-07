package text

import (
	"fmt"

	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/modules"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
	_ "github.com/go-go-golems/goja-text/pkg/extract"
	_ "github.com/go-go-golems/goja-text/pkg/markdown"
	_ "github.com/go-go-golems/goja-text/pkg/sanitize"
	helpdoc "github.com/go-go-golems/goja-text/pkg/xgoja/providers/text/doc"
)

const PackageID = "goja-text"

var textModuleNames = []string{
	"markdown",
	"sanitize",
	"extract",
}

// Register exposes goja-text modules as xgoja provider modules.
func Register(registry *providerapi.ProviderRegistry) error {
	entries := make([]providerapi.Entry, 0, len(textModuleNames))
	for _, name := range textModuleNames {
		mod := modules.GetModule(name)
		if mod == nil {
			return fmt.Errorf("text module %q is not registered", name)
		}
		entries = append(entries, nativeModuleEntry(mod))
	}
	entries = append(entries, providerapi.HelpSource{
		Name:        "runtime-api",
		Description: "goja-text Markdown, sanitize, and extract JavaScript API help pages",
		FS:          helpdoc.FS(),
		Root:        ".",
	})
	return registry.Package(PackageID, entries...)
}

func nativeModuleEntry(mod modules.NativeModule) providerapi.Module {
	return providerapi.Module{
		Name:        mod.Name(),
		DefaultAs:   mod.Name(),
		Description: mod.Doc(),
		NewModuleFactory: func(providerapi.ModuleSetupContext) (require.ModuleLoader, error) {
			return mod.Loader, nil
		},
	}
}

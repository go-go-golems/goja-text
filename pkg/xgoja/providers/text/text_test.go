package text

import (
	"testing"

	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
)

func TestRegisterExposesTypeScriptDescriptors(t *testing.T) {
	registry := providerapi.NewProviderRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register text provider: %v", err)
	}
	for _, name := range textModuleNames {
		mod, ok := registry.ResolveModule(PackageID, name)
		if !ok {
			t.Fatalf("missing module %s.%s", PackageID, name)
		}
		if mod.TypeScript == nil {
			t.Fatalf("expected module %s.%s to carry TypeScript descriptor", PackageID, name)
		}
	}
}

package chunking_test

import (
	"context"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/engine"
	_ "github.com/go-go-golems/goja-text/pkg/chunking"
)

func TestRequireChunkingSegmentsAndPacksGoBackedSpans(t *testing.T) {
	runtime := newChunkingRuntime(t)
	ret, err := runtime.Owner.Call(context.Background(), "chunking.integration", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
            const chunking = require("chunking");
            const source = "# Title\n\nFirst paragraph.\n\n## Detail\n\nSecond paragraph.\n";
            const blocks = chunking.markdownBlocks(source);
            const packed = chunking.pack(blocks.Spans, {
                maxUnits: 38,
                measure: "bytes",
                overlap: { unit: "spans", value: 1 },
                oversized: "allow",
            });
            const sections = chunking.markdownSections(source);
            ({
                sourcePreserved: blocks.Spans.map(span => span.Text).join("") === source,
                firstKind: blocks.Spans[0].Kind,
                firstStart: blocks.Spans[0].StartByte,
                chunks: packed.Chunks.length,
                firstWeight: packed.Chunks[0].Weight,
                sectionPath: sections.Spans[1].HeadingPath.join("/"),
                lowercaseMissing: blocks.spans === undefined,
            });
        `)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call: %v", err)
	}
	got := ret.(map[string]any)
	if got["sourcePreserved"] != true || got["firstKind"] != "heading" || got["firstStart"] != int64(0) || got["lowercaseMissing"] != true {
		t.Fatalf("projection = %#v", got)
	}
	if got["chunks"].(int64) < 2 || got["firstWeight"].(int64) <= 0 || got["sectionPath"] != "Title/Detail" {
		t.Fatalf("packing = %#v", got)
	}
}

func TestRequireChunkingWeightedAndRecursive(t *testing.T) {
	runtime := newChunkingRuntime(t)
	ret, err := runtime.Owner.Call(context.Background(), "chunking.weighted", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
            const chunking = require("chunking");
            const spans = chunking.lines("alpha\nbeta\ngamma\n").Spans;
            const weighted = chunking.packWeighted(
                spans.map((span, index) => ({ span, weight: index + 1 })),
                { maxWeight: 3, overlapWeight: 1 }
            );
            const recursive = chunking.recursive("abcdefghijklmnopqrstuvwxyz", {
                maxUnits: 5,
                measure: "runes",
                levels: ["lines", "runes"],
            });
            ({
                weightedChunks: weighted.Chunks.length,
                recursiveChunks: recursive.Chunks.length,
                recursiveMax: Math.max(...recursive.Chunks.map(chunk => chunk.Weight)),
                finalEnd: recursive.Chunks[recursive.Chunks.length - 1].EndByte,
            });
        `)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call: %v", err)
	}
	got := ret.(map[string]any)
	if got["weightedChunks"].(int64) < 2 || got["recursiveChunks"].(int64) != 6 || got["recursiveMax"].(int64) > 5 || got["finalEnd"].(int64) != 26 {
		t.Fatalf("result = %#v", got)
	}
}

func TestRequireChunkingRejectsUnknownAndMistypedOptions(t *testing.T) {
	runtime := newChunkingRuntime(t)
	for _, script := range []string{
		`require("chunking").lines("x", { keepTerminator: true })`,
		`require("chunking").pack([], { maxUnits: "5" })`,
	} {
		_, err := runtime.Owner.Call(context.Background(), "chunking.invalidOptions", func(_ context.Context, vm *goja.Runtime) (any, error) {
			_, runErr := vm.RunString(script)
			return nil, runErr
		})
		if err == nil || (!strings.Contains(err.Error(), "unknown option") && !strings.Contains(err.Error(), "must be an integer")) {
			t.Fatalf("script %q error = %v", script, err)
		}
	}
}

func newChunkingRuntime(t *testing.T) *engine.Runtime {
	t.Helper()
	factory, err := engine.NewRuntimeFactoryBuilder().UseModuleMiddleware(engine.MiddlewareOnly("chunking")).Build()
	if err != nil {
		t.Fatalf("build runtime factory: %v", err)
	}
	runtime, err := factory.NewRuntime(engine.WithStartupContext(context.Background()), engine.WithLifetimeContext(context.Background()))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close(context.Background()) })
	return runtime
}

package markdown_test

import (
	"context"
	"strings"
	"testing"

	"github.com/dop251/goja"
)

func TestRequireMarkdownDocumentParsesFrontmatterAndHeading(t *testing.T) {
	rt := newMarkdownRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "markdown.document.frontmatter", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const markdown = require("markdown");
			const source = [
				"---",
				"title: Demo slide",
				"number: \"01\"",
				"published: true",
				"---",
				"# Heading title",
				"",
				"Body text.",
			].join("\n");
			const doc = markdown.document(source)
				.Frontmatter().YAML().Repair().Optional().End()
				.Build();
			({
				title: doc.Frontmatter().String("title", "fallback"),
				number: doc.Frontmatter().String("number", "00"),
				published: doc.Frontmatter().Bool("published", false),
				missing: doc.Frontmatter().String("missing", "fallback"),
				heading: doc.FirstHeading("fallback"),
				body: doc.Body(),
				html: doc.RenderHTML(),
			});
		`)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call error = %v", err)
	}
	got, ok := ret.(map[string]any)
	if !ok {
		t.Fatalf("ret = %T, want map", ret)
	}
	if got["title"] != "Demo slide" || got["number"] != "01" || got["published"] != true || got["missing"] != "fallback" || got["heading"] != "Heading title" {
		t.Fatalf("unexpected document result: %#v", got)
	}
	if strings.Contains(got["body"].(string), "title: Demo slide") {
		t.Fatalf("body still contains frontmatter: %q", got["body"])
	}
	if !strings.Contains(got["html"].(string), "<h1>Heading title</h1>") {
		t.Fatalf("html = %q, want heading", got["html"])
	}
}

func TestRequireMarkdownDocumentExtractsAndStripsJSONBlock(t *testing.T) {
	rt := newMarkdownRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "markdown.document.blocks", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const markdown = require("markdown");
			const source = [
				"# Slide",
				"",
				"Visible paragraph.",
				"",
				String.fromCharCode(96, 96, 96) + "context-window",
				"{\"id\": \"demo\", \"parts\": [{\"id\": \"p1\"}],}",
				String.fromCharCode(96, 96, 96),
			].join("\n");
			const doc = markdown.document(source)
				.Blocks()
					.Block("context-window")
						.FromXMLTag("context-window")
						.FromFence("context-window")
						.JSON().Repair().Optional().End()
						.StripFromBody()
						.End()
					.End()
				.Build();
			const block = doc.Block("context-window");
			const json = block.JSONValue();
			({
				blockName: block.Name(),
				blockKind: block.Kind(),
				jsonId: json.id,
				jsonPart: json.parts[0].id,
				body: doc.Body(),
				html: doc.RenderHTML(),
			});
		`)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call error = %v", err)
	}
	got, ok := ret.(map[string]any)
	if !ok {
		t.Fatalf("ret = %T, want map", ret)
	}
	if got["blockName"] != "context-window" || got["blockKind"] != "fence" || got["jsonId"] != "demo" || got["jsonPart"] != "p1" {
		t.Fatalf("unexpected block result: %#v", got)
	}
	if strings.Contains(got["body"].(string), "context-window") || strings.Contains(got["html"].(string), "demo") {
		t.Fatalf("body/html still contains stripped block: %#v", got)
	}
}

func TestRequireMarkdownDocumentFrontmatterFieldSchemaAppliesDefaults(t *testing.T) {
	rt := newMarkdownRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "markdown.document.fieldSchema.defaults", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const markdown = require("markdown");
			const source = [
				"---",
				"title: Demo",
				"published: true",
				"---",
				"# Body",
			].join("\n");
			const doc = markdown.document(source)
				.Frontmatter().YAML().Optional()
					.Field("title").String().Required().End()
					.Field("number").String().Optional().Default("01").End()
					.Field("published").Bool().Required().End()
					.Field("weight").Number().Optional().Default(2.5).End()
					.End()
				.Build();
			({
				title: doc.Frontmatter().String("title", "fallback"),
				number: doc.Frontmatter().String("number", "00"),
				published: doc.Frontmatter().Bool("published", false),
				weight: doc.Frontmatter().Number("weight", 0),
				keys: doc.Frontmatter().Keys().join(","),
			});
		`)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call error = %v", err)
	}
	got, ok := ret.(map[string]any)
	if !ok {
		t.Fatalf("ret = %T, want map", ret)
	}
	if got["title"] != "Demo" || got["number"] != "01" || got["published"] != true || got["weight"] != 2.5 {
		t.Fatalf("unexpected schema/default result: %#v", got)
	}
	if got["keys"] != "number,published,title,weight" {
		t.Fatalf("keys = %#v", got["keys"])
	}
}

func TestRequireMarkdownDocumentFrontmatterFieldSchemaRejectsMissingRequired(t *testing.T) {
	rt := newMarkdownRuntime(t)

	_, err := rt.Owner.Call(context.Background(), "markdown.document.fieldSchema.required", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, runErr := vm.RunString(`
			const markdown = require("markdown");
			markdown.document("---\ntitle: Demo\n---\n# Body")
				.Frontmatter().YAML().Optional()
					.Field("id").String().Required().End()
					.End()
				.Build();
		`)
		return nil, runErr
	})
	if err == nil || !strings.Contains(err.Error(), "frontmatter field \"id\": required field missing") {
		t.Fatalf("error = %v, want missing required field", err)
	}
}

func TestRequireMarkdownDocumentFrontmatterFieldSchemaRejectsTypeMismatch(t *testing.T) {
	rt := newMarkdownRuntime(t)

	_, err := rt.Owner.Call(context.Background(), "markdown.document.fieldSchema.type", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, runErr := vm.RunString(`
			const markdown = require("markdown");
			markdown.document("---\npublished: \"yes\"\n---\n# Body")
				.Frontmatter().YAML().Optional()
					.Field("published").Bool().Required().End()
					.End()
				.Build();
		`)
		return nil, runErr
	})
	if err == nil || !strings.Contains(err.Error(), "frontmatter field \"published\": expected bool") {
		t.Fatalf("error = %v, want type mismatch", err)
	}
}

func TestRequireMarkdownDocumentRejectsInvalidBuilderConfig(t *testing.T) {
	rt := newMarkdownRuntime(t)

	_, err := rt.Owner.Call(context.Background(), "markdown.document.validation", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, runErr := vm.RunString(`
			const markdown = require("markdown");
			markdown.document("# Bad")
				.Blocks().Block("context window").FromFence("context-window").End().End()
				.Build();
		`)
		return nil, runErr
	})
	if err == nil || !strings.Contains(err.Error(), "invalid block name") {
		t.Fatalf("error = %v, want invalid block name", err)
	}
}

func TestRequireMarkdownDocumentRequiredBlockError(t *testing.T) {
	rt := newMarkdownRuntime(t)

	_, err := rt.Owner.Call(context.Background(), "markdown.document.requiredBlock", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, runErr := vm.RunString(`
			const markdown = require("markdown");
			markdown.document("# Missing")
				.Blocks()
					.Block("context-window").FromFence("context-window").Required().End()
					.End()
				.Build();
		`)
		return nil, runErr
	})
	if err == nil || !strings.Contains(err.Error(), "required block \"context-window\" not found") {
		t.Fatalf("error = %v, want required block error", err)
	}
}

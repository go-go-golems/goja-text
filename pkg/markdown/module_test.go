package markdown_test

import (
	"context"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/engine"
	md "github.com/go-go-golems/goja-text/pkg/markdown"
)

func TestRequireMarkdownParseExposesGoFields(t *testing.T) {
	rt := newMarkdownRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "markdown.parse.fields", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const markdown = require("markdown");
			const ast = markdown.parse("# Hello\n\nWorld");
			({
				type: ast.Type,
					firstType: ast.Children[0].Type,
					firstLevel: ast.Children[0].Level,
					firstText: ast.Children[0].Children[0].Text,
					firstStartByte: ast.Children[0].StartByte,
					firstEndByte: ast.Children[0].EndByte,
					firstStartRune: ast.Children[0].StartRune,
					firstEndRune: ast.Children[0].EndRune,
					firstEndLine: ast.Children[0].EndLine,
					firstEndColumn: ast.Children[0].EndColumn,
					lowercaseMissing: ast.type === undefined,
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
		t.Fatalf("ret = %T, want map[string]any", ret)
	}
	if got["type"] != "document" || got["firstType"] != "heading" || got["firstText"] != "Hello" || got["lowercaseMissing"] != true {
		t.Fatalf("unexpected JS projection result: %#v", got)
	}
	if got["firstLevel"] != int64(1) && got["firstLevel"] != int(1) && got["firstLevel"] != float64(1) {
		t.Fatalf("firstLevel = %#v, want 1", got["firstLevel"])
	}
	for key, want := range map[string]int64{
		"firstStartByte": 0,
		"firstEndByte":   7,
		"firstStartRune": 0,
		"firstEndRune":   7,
		"firstEndLine":   1,
		"firstEndColumn": 8,
	} {
		if got[key] != want {
			t.Fatalf("%s = %#v, want %d", key, got[key], want)
		}
	}
}

func TestRequireMarkdownExposesGoldmarkEdgeFieldsToJS(t *testing.T) {
	rt := newMarkdownRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "markdown.parse.edgeFields", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const markdown = require("markdown");
			const ast = markdown.parse([
				"![Alt *em* text](https://img.example/p.png \"Image Title\")",
				"",
				"~~~go title=\"demo\"",
				"fmt.Println(1)",
				"~~~",
				"",
			].join("\n"));
			const seen = { image: null, fenced: null };
			markdown.walk(ast, (node) => {
				if (node.Type === "image") seen.image = {
					Destination: node.Destination,
					Title: node.Title,
					Alt: node.Alt,
					StartLine: node.StartLine,
					StartColumn: node.StartColumn,
					SourcePos0: node.SourcePos[0],
					SourcePos1: node.SourcePos[1],
				};
				if (node.Type === "fencedCodeBlock") seen.fenced = {
					Language: node.Language,
					Info: node.Info,
					TextContainsPrint: node.Text.includes("fmt.Println"),
					StartLine: node.StartLine,
					StartColumn: node.StartColumn,
					SourcePos0: node.SourcePos[0],
					SourcePos1: node.SourcePos[1],
				};
			});
			seen;
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
		t.Fatalf("ret = %T, want map[string]any", ret)
	}
	image, ok := got["image"].(map[string]any)
	if !ok {
		t.Fatalf("image = %#v, want map", got["image"])
	}
	if image["Destination"] != "https://img.example/p.png" || image["Title"] != "Image Title" || image["Alt"] != "Alt em text" {
		t.Fatalf("image fields = %#v", image)
	}
	if image["StartLine"] != int64(1) || image["StartColumn"] != int64(1) {
		t.Fatalf("image source fields = %#v, want 1:1", image)
	}
	fenced, ok := got["fenced"].(map[string]any)
	if !ok {
		t.Fatalf("fenced = %#v, want map", got["fenced"])
	}
	if fenced["Language"] != "go" || fenced["Info"] != "go title=\"demo\"" || fenced["TextContainsPrint"] != true {
		t.Fatalf("fenced fields = %#v", fenced)
	}
	if fenced["StartLine"] != int64(3) || fenced["StartColumn"] != int64(1) {
		t.Fatalf("fenced source fields = %#v, want 3:1", fenced)
	}
}

func TestRequireMarkdownWalkSupportsJSQueries(t *testing.T) {
	rt := newMarkdownRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "markdown.walk.query", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const markdown = require("markdown");
			const ast = markdown.parse("# H1\n\nSee [site](https://example.com).\n\n## H2");
			const headings = [];
			const links = [];
			markdown.walk(ast, (node, ctx) => {
				if (node.Type === "heading") headings.push({ level: node.Level, text: markdown.textContent(node), depth: ctx.Depth });
				if (node.Type === "link") links.push({ destination: node.Destination, text: markdown.textContent(node) });
			});
			({ headings, links });
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
		t.Fatalf("ret = %T, want map[string]any", ret)
	}
	headings, ok := got["headings"].([]any)
	if !ok || len(headings) != 2 {
		t.Fatalf("headings = %#v, want two headings", got["headings"])
	}
	links, ok := got["links"].([]any)
	if !ok || len(links) != 1 {
		t.Fatalf("links = %#v, want one link", got["links"])
	}
}

func TestValidateRejectsInvalidNode(t *testing.T) {
	rt := newMarkdownRuntime(t)
	_ = rt.VM.Set("invalidNode", &md.MarkdownNode{Type: "heading", Level: 99})

	ret, err := rt.Owner.Call(context.Background(), "markdown.validate.node", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const markdown = require("markdown");
			markdown.validate(invalidNode);
		`)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call error = %v", err)
	}
	result, ok := ret.(md.ValidationResult)
	if !ok {
		t.Fatalf("ret = %T, want markdown.ValidationResult", ret)
	}
	if result.Valid || len(result.Errors) == 0 {
		t.Fatalf("result = %#v, want invalid with errors", result)
	}
}

func TestRequireMarkdownBuilderRendersDocument(t *testing.T) {
	rt := newMarkdownRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "markdown.builder.document", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const markdown = require("markdown");
			const result = markdown.builder()
				.Title("Sprint report")
				.Paragraph("Generated from JS data.")
				.Table()
					.Columns({ label: "Name", align: "left" }, { label: "Status", align: "right" })
					.Row("Parser", "done")
					.Row("Builder", "planned")
					.End()
				.Heading(2, "Next steps")
				.Checklist([{ text: "Expose goja API", checked: true }, { text: "Write docs" }])
				.Render();
			({ text: result.Text, bytes: result.Bytes, blocks: result.Blocks });
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
	text, ok := got["text"].(string)
	if !ok {
		t.Fatalf("text = %T", got["text"])
	}
	for _, want := range []string{"# Sprint report", "| Name    | Status  |", "| :------ | ------: |", "- [x] Expose goja API", "- [ ] Write docs"} {
		if !strings.Contains(text, want) {
			t.Fatalf("text missing %q:\n%s", want, text)
		}
	}
}

func TestRequireMarkdownInlineHelpersAndRenderHTML(t *testing.T) {
	rt := newMarkdownRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "markdown.builder.inline", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const markdown = require("markdown");
			const i = markdown.inline();
			const builder = markdown.builder()
				.Paragraph("Run ", i.Code("go test ./..."), " and read ", i.Link("docs", "https://example.com"), ".")
				.CodeBlock("js", "console.log('ok')");
			({ md: builder.RenderString(), html: builder.RenderHTML() });
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
	mdText := got["md"].(string)
	if !strings.Contains(mdText, "`go test ./...`") || !strings.Contains(mdText, "[docs](https://example.com)") {
		t.Fatalf("markdown text = %s", mdText)
	}
	html := got["html"].(string)
	if !strings.Contains(html, "<code>go test ./...</code>") || !strings.Contains(html, "<pre><code class=\"language-js\">") {
		t.Fatalf("html = %s", html)
	}
}

func TestRequireMarkdownBuilderValidationError(t *testing.T) {
	rt := newMarkdownRuntime(t)

	_, err := rt.Owner.Call(context.Background(), "markdown.builder.validation", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, runErr := vm.RunString(`
			const markdown = require("markdown");
			markdown.builder()
				.Heading(9, "bad")
				.Table().Columns("A", "B").Row("one").End()
				.RenderString();
		`)
		return nil, runErr
	})
	if err == nil || !strings.Contains(err.Error(), "heading level") || !strings.Contains(err.Error(), "table row") {
		t.Fatalf("error = %v, want builder validation error", err)
	}
}

func newMarkdownRuntime(t *testing.T) *engine.Runtime {
	t.Helper()
	factory, err := engine.NewRuntimeFactoryBuilder().UseModuleMiddleware(engine.MiddlewareOnly("markdown")).Build()
	if err != nil {
		t.Fatalf("build runtime factory: %v", err)
	}
	rt, err := factory.NewRuntime(engine.WithStartupContext(context.Background()), engine.WithLifetimeContext(context.Background()))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	t.Cleanup(func() { _ = rt.Close(context.Background()) })
	return rt
}

package template

import (
	"bytes"
	fmt "fmt"
	htmltemplate "html/template"
	"sort"
	texttemplate "text/template"
)

type TemplateSet struct {
	Mode Mode
	Name string
	text *texttemplate.Template
	html *htmltemplate.Template
}

func parseTemplateSet(cfg TemplateConfig, name, source string) (*TemplateSet, error) {
	funcs, err := funcMapFor(cfg.Mode, cfg.FuncSets)
	if err != nil {
		return nil, err
	}
	option := "missingkey=" + cfg.MissingKey
	set := &TemplateSet{Mode: cfg.Mode, Name: name}
	if cfg.Mode == ModeHTML {
		t := htmltemplate.New(cfg.Name).Funcs(funcs).Option(option)
		if cfg.LeftDelim != "" {
			t = t.Delims(cfg.LeftDelim, cfg.RightDelim)
		}
		parsed, err := t.New(name).Parse(source)
		if err != nil {
			return nil, fmt.Errorf("template.html.parse: %w", err)
		}
		set.html = parsed
		return set, nil
	}
	t := texttemplate.New(cfg.Name).Funcs(funcs).Option(option)
	if cfg.LeftDelim != "" {
		t = t.Delims(cfg.LeftDelim, cfg.RightDelim)
	}
	parsed, err := t.New(name).Parse(source)
	if err != nil {
		return nil, fmt.Errorf("template.text.parse: %w", err)
	}
	set.text = parsed
	return set, nil
}

func (s *TemplateSet) Render(data any) (*RenderResult, error) {
	return s.RenderTemplate(s.Name, data)
}

func (s *TemplateSet) RenderString(data any) (string, error) {
	result, err := s.Render(data)
	if err != nil {
		return "", err
	}
	return result.Text, nil
}

func (s *TemplateSet) RenderTemplate(name string, data any) (*RenderResult, error) {
	if s == nil {
		return nil, fmt.Errorf("template.render: template set is nil")
	}
	if name == "" {
		name = s.Name
	}
	var buf bytes.Buffer
	if s.Mode == ModeHTML {
		if s.html == nil {
			return nil, fmt.Errorf("template.html.render: template set is not initialized")
		}
		if err := s.html.ExecuteTemplate(&buf, name, data); err != nil {
			return nil, fmt.Errorf("template.html.render %q: %w", name, err)
		}
	} else {
		if s.text == nil {
			return nil, fmt.Errorf("template.text.render: template set is not initialized")
		}
		if err := s.text.ExecuteTemplate(&buf, name, data); err != nil {
			return nil, fmt.Errorf("template.text.render %q: %w", name, err)
		}
	}
	out := buf.String()
	return &RenderResult{Text: out, TemplateName: name, Mode: s.Mode, Bytes: len([]byte(out))}, nil
}

func (s *TemplateSet) Templates() []TemplateInfo {
	if s == nil {
		return nil
	}
	var ret []TemplateInfo
	if s.Mode == ModeHTML && s.html != nil {
		for _, tmpl := range s.html.Templates() {
			ret = append(ret, TemplateInfo{Name: tmpl.Name(), Defined: tmpl.Tree != nil, Mode: s.Mode})
		}
	} else if s.text != nil {
		for _, tmpl := range s.text.Templates() {
			ret = append(ret, TemplateInfo{Name: tmpl.Name(), Defined: tmpl.Tree != nil, Mode: s.Mode})
		}
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Name < ret[j].Name })
	return ret
}

func (s *TemplateSet) Lookup(name string) *TemplateInfo {
	for _, info := range s.Templates() {
		if info.Name == name {
			v := info
			return &v
		}
	}
	return nil
}

func RenderText(source string, data any) (*RenderResult, error) {
	set, err := NewTextBuilder().Parse(source)
	if err != nil {
		return nil, err
	}
	return set.Render(data)
}

func RenderHTML(source string, data any) (*RenderResult, error) {
	set, err := NewHTMLBuilder().Parse(source)
	if err != nil {
		return nil, err
	}
	return set.Render(data)
}

package service

import (
	"bytes"
	"errors"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type MarkdownService struct {
	parser goldmark.Markdown
}

type ConvertToCSDNInput struct {
	Title          string
	Markdown       string
	IncludeWrapper bool
}

type ConvertToCSDNOutput struct {
	Format      string `json:"format"`
	Title       string `json:"title,omitempty"`
	HTML        string `json:"html"`
	WrappedHTML string `json:"wrapped_html,omitempty"`
}

func NewMarkdownService() *MarkdownService {
	return &MarkdownService{
		parser: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
			),
		),
	}
}

func (s *MarkdownService) ConvertToCSDN(input *ConvertToCSDNInput) (*ConvertToCSDNOutput, error) {
	if input == nil {
		return nil, errors.New("转换参数不能为空")
	}

	source := strings.TrimSpace(input.Markdown)
	if source == "" {
		return nil, errors.New("Markdown 内容不能为空")
	}

	var buf bytes.Buffer
	if err := s.parser.Convert([]byte(source), &buf); err != nil {
		return nil, err
	}

	convertedHTML := strings.TrimSpace(buf.String())
	output := &ConvertToCSDNOutput{
		Format: "csdn-html",
		Title:  strings.TrimSpace(input.Title),
		HTML:   convertedHTML,
	}

	if input.IncludeWrapper {
		output.WrappedHTML = wrapCSDNHTML(convertedHTML)
	}

	return output, nil
}

func wrapCSDNHTML(content string) string {
	return `<div id="content_views" class="markdown_views prism-atom-one-dark">` + "\n" + content + "\n" + `</div>`
}

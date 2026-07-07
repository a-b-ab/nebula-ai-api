package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkdownServiceConvertToCSDN(t *testing.T) {
	service := NewMarkdownService()

	markdown, err := os.ReadFile("testdata/csdn_sample.md")
	if err != nil {
		t.Fatalf("read fixture markdown: %v", err)
	}

	result, err := service.ConvertToCSDN(&ConvertToCSDNInput{
		Title:          "测试文章",
		Markdown:       string(markdown),
		IncludeWrapper: true,
	})
	if err != nil {
		t.Fatalf("ConvertToCSDN returned error: %v", err)
	}

	assertContains(t, result.HTML, `<h1>Markdown 转 CSDN 示例</h1>`)
	assertContains(t, result.HTML, `<li>保留标题、列表、引用、表格和代码块</li>`)
	assertContains(t, result.HTML, `<code class="language-go">`)
	assertContains(t, result.HTML, `<table>`)
	assertContains(t, result.HTML, `<blockquote>`)
	assertContains(t, result.WrappedHTML, `id="content_views"`)
	assertContains(t, result.WrappedHTML, `markdown_views prism-atom-one-dark`)
}

func TestMarkdownServiceConvertExternalMarkdownFileToCSDN(t *testing.T) {
	mdPath := os.Getenv("CSDN_MD_FILE")
	if mdPath == "" {
		t.Skip("set CSDN_MD_FILE=/path/to/article.md to generate CSDN HTML from an external file")
	}

	markdown, err := readMarkdownFile(mdPath)
	if err != nil {
		t.Fatalf("read external markdown file %q: %v", mdPath, err)
	}

	title := os.Getenv("CSDN_TITLE")
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(mdPath), filepath.Ext(mdPath))
	}

	result, err := NewMarkdownService().ConvertToCSDN(&ConvertToCSDNInput{
		Title:          title,
		Markdown:       string(markdown),
		IncludeWrapper: true,
	})
	if err != nil {
		t.Fatalf("ConvertToCSDN returned error: %v", err)
	}

	articleHTML := result.WrappedHTML
	if articleHTML == "" {
		articleHTML = result.HTML
	}

	outPath := os.Getenv("CSDN_HTML_OUT")
	if outPath == "" {
		t.Logf("CSDN HTML:\n%s", articleHTML)
		return
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		t.Fatalf("create output dir for %q: %v", outPath, err)
	}
	if err := os.WriteFile(outPath, []byte(articleHTML), 0o644); err != nil {
		t.Fatalf("write CSDN HTML to %q: %v", outPath, err)
	}

	t.Logf("CSDN HTML written to %s", outPath)
}

func TestMarkdownServiceConvertToCSDNRejectsEmptyMarkdown(t *testing.T) {
	service := NewMarkdownService()

	_, err := service.ConvertToCSDN(&ConvertToCSDNInput{
		Markdown: "   ",
	})
	if err == nil {
		t.Fatal("expected empty Markdown to return an error")
	}
}

func assertContains(t *testing.T, value, expected string) {
	t.Helper()
	if !strings.Contains(value, expected) {
		t.Fatalf("expected %q to contain %q", value, expected)
	}
}

func readMarkdownFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err == nil || filepath.IsAbs(path) || !errors.Is(err, os.ErrNotExist) {
		return content, err
	}

	return os.ReadFile(filepath.Join("..", "..", path))
}

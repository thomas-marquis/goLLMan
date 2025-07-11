package components

import (
	"bytes"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"strings"
)

var mdCConfig = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.DefinitionList,
		extension.Footnote,
		extension.Typographer,
		extension.Linkify,
		extension.Strikethrough,
		extension.Table,
		extension.TaskList,
		emoji.Emoji,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
		parser.WithAttribute(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
		html.WithUnsafe(),
	),
)

func convertToHTML(content string) (string, error) {
	var buf bytes.Buffer

	if err := mdCConfig.Convert([]byte(content), &buf); err != nil {
		return "", err
	}

	htmlContent := strings.ReplaceAll(buf.String(), "<pre>",
		"<pre class=\"overflow-x-auto p-4 rounded\">")

	return htmlContent, nil
}

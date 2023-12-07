package chatgpt

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
)

func SummaryTemplateBuilder(input *SummaryTemplateInput) (string, error) {
	tmpl, err := template.New("summary").Parse(gptRequestSummaryTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, input); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buffer.String(), nil
}

//go:embed summary_template.txt
var gptRequestSummaryTemplate string

type SummaryTemplateInput struct {
	Title   string
	Content string
}

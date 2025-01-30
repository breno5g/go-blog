package helpers

import (
	"html/template"
	"os"
	"strings"

	"github.com/yuin/goldmark"
)

func ConvertMDtoHTML(inputPath string) (template.HTML, error) {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return "", err
	}

	var htmlContent strings.Builder
	goldmark.Convert(content, &htmlContent)

	return template.HTML(htmlContent.String()), nil
}

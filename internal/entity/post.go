package entity

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/breno5g/go-blog/config"
	"github.com/breno5g/go-blog/internal/helpers"
)

type Post struct {
	Title    string
	Content  template.HTML
	Filename string
}

type Posts struct {
	Posts     []Post
	InputPath string
	OutputDir string
	Logger    *config.Logger
}

func (p *Posts) Set() {
	files, err := os.ReadDir(p.InputPath)
	if err != nil {
		p.Logger.Errorf("Error reading directory: %v", err)
		panic(err)
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
			continue
		}

		inputPath := fmt.Sprintf("%s/%s", p.InputPath, f.Name())
		outputFilename := strings.TrimSuffix(f.Name(), ".md") + ".html"

		content, err := helpers.ConvertMDtoHTML(inputPath)
		if err != nil {
			p.Logger.Errorf("Error converting MD to HTML: %v", err)
			panic(err)
		}

		p.Posts = append(p.Posts, Post{
			Title:    strings.TrimSuffix(f.Name(), ".md"),
			Content:  content,
			Filename: outputFilename,
		})
	}
}

func (p *Posts) GetAll() []Post {
	return p.Posts
}

package entity

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
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
	Posts         []Post
	InputPath     string
	OutputDir     string
	Logger        *config.Logger
	PostTemplate  *template.Template
	IndexTemplate *template.Template
}

func (p *Post) Create(outputDir, outputFilename string, postTemplate *template.Template) error {
	outputPath := filepath.Join(outputDir, outputFilename)
	outFile, err := os.Create(outputPath)

	if err != nil {
		return err
	}

	defer outFile.Close()

	return postTemplate.Execute(outFile, p)
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

		post := Post{
			Title:    strings.TrimSuffix(f.Name(), ".md"),
			Content:  content,
			Filename: outputFilename,
		}

		p.Posts = append(p.Posts, post)

		post.Create(p.OutputDir, outputFilename, p.PostTemplate)
	}
}

func (p *Posts) BuildIndex() {
	indexFile, err := os.Create(filepath.Join(p.OutputDir, "index.html"))
	if err != nil {
		p.Logger.Errorf("Error creating index file: %v", err)
		panic(err)
	}

	defer indexFile.Close()

	if err := p.IndexTemplate.Execute(indexFile, p.Posts); err != nil {
		p.Logger.Errorf("Error executing index template: %v", err)
		panic(err)
	}
}

func (p *Posts) GetAll() []Post {
	return p.Posts
}

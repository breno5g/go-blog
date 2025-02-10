package main

import (
	"html/template"
	"os"
	"path/filepath"

	"github.com/breno5g/go-blog/config"
	"github.com/breno5g/go-blog/internal/entity"
)

type Post struct {
	Title    string
	Content  template.HTML
	Filename string
}

// var (
// 	templatesDir = "templates"

// 	postTemplate  = template.Must(template.ParseFiles(filepath.Join(templatesDir, "post.html")))
// 	indexTemplate = template.Must(template.ParseFiles(filepath.Join(templatesDir, "index.html")))
// )

func init() {
	outputDir := config.GetPaths().OutputDir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}
}

func main() {
	paths := config.GetPaths()

	posts := entity.Posts{
		InputPath:     paths.PostsDir,
		OutputDir:     paths.OutputDir,
		Logger:        config.GetLogger("POSTS"),
		PostTemplate:  template.Must(template.ParseFiles(filepath.Join(paths.TemplateDir, "post.html"))),
		IndexTemplate: template.Must(template.ParseFiles(filepath.Join(paths.TemplateDir, "index.html"))),
	}

	posts.Set()
	posts.BuildIndex()

}

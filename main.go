package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type Post struct {
	Title    string
	Content  template.HTML
	Filename string
}

var (
	postsDir     = "posts"
	outputDir    = "public"
	templatesDir = "templates"

	postTemplate  = template.Must(template.ParseFiles(filepath.Join(templatesDir, "post.html")))
	indexTemplate = template.Must(template.ParseFiles(filepath.Join(templatesDir, "index.html")))
)

func init() {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}
}

func main() {
	var posts []Post
	files, err := os.ReadDir(postsDir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
			continue
		}

		post := Post{
			Title:    strings.TrimSuffix(f.Name(), ".md"),
			Filename: filepath.Join(postsDir, f.Name()),
		}

		posts = append(posts, post)
	}

	for _, post := range posts {
		fmt.Println(post.Filename)
	}
}

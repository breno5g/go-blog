package main

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
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

func convertMDtoHTML(inputPath string) (template.HTML, error) {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return "", err
	}

	var htmlContent strings.Builder
	goldmark.Convert(content, &htmlContent)

	return template.HTML(htmlContent.String()), nil
}

func sortFilesByCreationTime(files []os.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})
}

func main() {
	var posts []Post
	files, err := ioutil.ReadDir(postsDir)
	if err != nil {
		panic(err)
	}

	sortFilesByCreationTime(files)

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
			continue
		}

		inputPath := filepath.Join(postsDir, f.Name())

		content, err := convertMDtoHTML(inputPath)

		if err != nil {
			panic(err)
		}
		outputFilename := strings.TrimSuffix(f.Name(), ".md") + ".html"

		post := Post{
			Title:    strings.TrimSuffix(f.Name(), ".md"),
			Content:  content,
			Filename: outputFilename,
		}

		posts = append(posts, post)

		outputPath := filepath.Join(outputDir, outputFilename)
		outFile, _ := os.Create(outputPath)
		defer outFile.Close()
		postTemplate.Execute(outFile, post)
	}

	indexFile, err := os.Create(filepath.Join(outputDir, "index.html"))

	if err != nil {
		panic(err)
	}

	defer indexFile.Close()
	indexTemplate.Execute(indexFile, posts)

}

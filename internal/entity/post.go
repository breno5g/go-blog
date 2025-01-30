package entity

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/breno5g/go-blog/config"
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

		fmt.Println(f.Name())
	}
}

func (p *Posts) GetAll() []Post {
	return p.Posts
}

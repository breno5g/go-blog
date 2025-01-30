package config

import (
	"os"
)

type Paths struct {
	PostsDir    string
	OutputDir   string
	TemplateDir string
}

func InitilizeConstants() Paths {
	postsDir := os.Getenv("POSTS_DIR")
	outputsDir := os.Getenv("OUTPUT_DIR")
	templatesDir := os.Getenv("TEMPLATE_DIR")

	if postsDir == "" {
		postsDir = "posts"
	}

	if outputsDir == "" {
		outputsDir = "public"
	}

	if templatesDir == "" {
		templatesDir = "templates"
	}

	paths := Paths{
		PostsDir:    postsDir,
		OutputDir:   outputsDir,
		TemplateDir: templatesDir,
	}

	return paths
}

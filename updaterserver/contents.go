package main

import (
	"os"
	"path/filepath"
)

type ContentsInfo struct {
	FilePath	string
	FileInfo	os.FileInfo
}

type Contents struct {
	RootPath 	string
	AllContents []ContentsInfo
}

var (
	handler Contents
)

func GetContents() *Contents {
	return &handler
}

func (c *Contents) LoadContents(rootPath string) {
	if c.RootPath == "" {
		c.RootPath = rootPath
	}

	_ = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			ci := ContentsInfo{FilePath: path, FileInfo: info}
			c.AllContents = append(c.AllContents, ci)
		}

		return nil
	})
}

func (c *Contents) Reload() {
	c.AllContents = nil
	c.LoadContents(c.RootPath)
}
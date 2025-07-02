package internal

import (
	"embed"
	"path"
)

//go:embed assets/*
var assets embed.FS

func GetAsset(file string) []byte {
	data, err := assets.ReadFile(path.Join("assets", file))
	if err != nil {
		panic(err)
	}
	return data
}

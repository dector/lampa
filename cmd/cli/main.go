package main

import (
	"lampa/internal"
	"os"
)

func main() {
	file := ""
	configName := ""
	if len(os.Args) > 2 {
		file = os.Args[1]
		configName = os.Args[2]
	} else {
		os.Stderr.WriteString("usage: lampa <file> <configuration name>\n")
		os.Exit(1)
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		os.Stderr.WriteString("file does not exist\n")
		os.Exit(1)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	content := string(data)

	internal.ParseTreeFromOutput(content, configName)
}

package main

import (
	"fmt"
	"os"

	. "lampa/internal/globals"

	"github.com/samber/lo"
)

func main() {
	command := lo.NthOrEmpty(os.Args, 1)

	switch command {
	case "collect":
		execCollect()
	case "help":
		printUsage()
		os.Exit(0)
	case "version":
		fmt.Println(G.Version)
		os.Exit(0)
	default:
		printUsage()
		os.Exit(2)
	}

	// file := ""
	// configName := ""
	// if len(os.Args) > 2 {
	// 	file = os.Args[1]
	// 	configName = os.Args[2]
	// } else {
	// 	os.Stderr.WriteString("usage: lampa <file> <configuration name>\n")
	// 	os.Exit(1)
	// }
	// if _, err := os.Stat(file); os.IsNotExist(err) {
	// 	os.Stderr.WriteString("file does not exist\n")
	// 	os.Exit(1)
	// }

	// data, err := os.ReadFile(file)
	// if err != nil {
	// 	panic(err)
	// }
	// content := string(data)

	// tree, err := internal.ParseTreeFromOutput(content, configName)
	// if err != nil {
	// 	panic(err)
	// }

	// // jsonTree, err := json.MarshalIndent(tree.Summary, "", "  ")
	// // if err != nil {
	// // 	panic(err)
	// // }

	// for _, it := range tree.Summary {
	// 	fmt.Println(it.String())
	// }
}

func execCollect() {
	fmt.Println("Not Implemented Yet")
	os.Exit(127)
}

func printUsage() {
	fmt.Println("usage: lampa ...")
}

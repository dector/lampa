package main

import (
	"context"
	"fmt"
	"log"
	"os"

	. "lampa/internal/globals"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name: "lampa",
		Commands: []*cli.Command{
			{
				Name: "collect",
				Action: func(ctx context.Context, c *cli.Command) error {
					execCollect()
					return nil
				},
			},
			{
				Name: "version",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println(G.Version)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
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

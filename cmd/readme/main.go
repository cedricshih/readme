package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"readme/api/readme"
)

const (
	localConfig = "local.yaml"
)

type arguments struct {
	apiKey    string
	docRoot   string
	rawOutput bool
	allYes    bool
}

var args = &arguments{}

func usage(msg string) {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "%s [args...] <command>\n", filepath.Base(os.Args[0]))
	flag.Usage()
	if msg != "" {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
	}
}

func main() {
	flag.StringVar(&args.apiKey, "k", args.apiKey, "API Key")
	flag.StringVar(&args.docRoot, "d", args.docRoot, "Document folder")
	flag.BoolVar(&args.rawOutput, "j", args.rawOutput, "Output JSON response")
	flag.BoolVar(&args.allYes, "y", args.allYes, "'Yes' to all prompts")
	flag.Parse()
	cfg, err := ReadLocalConfig(localConfig)
	if err != nil {
		panic(err)
	}
	if cfg.APIKey != "" {
		args.apiKey = cfg.APIKey
	}
	if cfg.DocRoot != "" {
		args.docRoot = cfg.DocRoot
	}
	if args.apiKey == "" {
		usage("API Key is not specified either by command line argument or in 'local.yaml'")
		os.Exit(-1)
	}
	if args.docRoot == "" {
		usage("Document folder is not specified either by command line argument or in 'local.yaml'")
		os.Exit(-1)
	}
	t := &tool{
		input:   os.Stdin,
		output:  os.Stdout,
		client:  readme.NewClient(args.apiKey),
		docRoot: args.docRoot,
		allYes:  args.allYes,
	}
	if args.rawOutput {
		t.client.Output = os.Stdout
	}
	cmd := flag.Arg(0)
	switch cmd {
	case "cat", "category":
		err = t.categories()
	case "docs":
		err = t.docs(flag.Arg(1))
	case "doc":
		err = t.doc(flag.Arg(1))
	case "pull":
		err = t.pull(flag.Arg(1))
	case "push":
		err = t.push(flag.Arg(1))
	default:
		usage("")
	}
	if err != nil {
		log.Printf("Command '%s' failed: %s", cmd, err.Error())
	}
}

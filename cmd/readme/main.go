package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/cedricshih/readme/api/readme"
)

const (
	localConfig = "local.yaml"
)

var args = &struct {
	apiKey    string
	help      bool
	rawOutput bool
}{}

var remoteCommand = &RemoteCommand{
	input:  os.Stdin,
	output: os.Stdout,
}

var commands = map[string]Command{
	// "cat":        &GetCategory{RemoteCommand: rc},
	// "cats":       &ListCategories{RemoteCommand: rc},
	// "categories": &ListCategories{RemoteCommand: rc},
	// "docs":       &ListDocuments{RemoteCommand: rc},
	// "doc":        &GetDocument{RemoteCommand: rc},
	"pull": &PullDocument{remoteCommand},
	// "push":       &PushDocument{RemoteCommand: rc},
	// "clone":      &PushDocument{RemoteCommand: rc},
}

func usage(w io.Writer, fmtsrt string, args ...interface{}) {
	fmt.Fprintf(w, "%s [args...] <command>\n", filepath.Base(os.Args[0]))
	flag.Usage()
	names := []string{}
	for k := range commands {
		names = append(names, k)
	}
	fmt.Fprintf(w, "Available commands: %s\n", strings.Join(names, ", "))
	if fmtsrt != "" {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(fmtsrt, args...))
	}
}

func main() {
	flag.BoolVar(&args.help, "h", args.help, "help")
	flag.StringVar(&args.apiKey, "k", args.apiKey, "API Key")
	flag.StringVar(&remoteCommand.docRoot, "d", remoteCommand.docRoot, "Document folder")
	flag.BoolVar(&args.rawOutput, "j", args.rawOutput, "Output JSON response")
	flag.BoolVar(&remoteCommand.allYes, "y", remoteCommand.allYes, "'Yes' to all prompts")
	flag.Parse()
	out := flag.CommandLine.Output()
	cmdname := flag.Arg(0)
	if args.help && cmdname == "" {
		usage(out, "")
		return
	}
	cfg, err := ReadLocalConfig(localConfig)
	if err != nil {
		panic(err)
	}
	if cfg.APIKey != "" {
		log.Printf("Using API key from %s: %s", localConfig, cfg.APIKey)
		args.apiKey = cfg.APIKey
	}
	if cfg.DocRoot != "" {
		log.Printf("Using doc root from %s: %s", localConfig, cfg.DocRoot)
		remoteCommand.docRoot = cfg.DocRoot
	}
	if args.apiKey == "" {
		usage(out, "API Key is not specified either by command line argument or in 'local.yaml'")
		os.Exit(int(syscall.EINVAL))
	}
	if cmdname == "" {
		usage(out, "Missing command")
		os.Exit(int(syscall.EINVAL))
	}
	remoteCommand.client = readme.NewClient(args.apiKey)
	if args.rawOutput {
		remoteCommand.client.Output = os.Stdout
	}
	cmd := commands[cmdname]
	if cmd == nil {
		usage(out, "Unknown command: %s", cmdname)
		os.Exit(int(syscall.EINVAL))
	}
	if args.help {
		cmd.Usage(out, os.Args[0], flag.Arg(0))
		return
	}
	if len(flag.Args())-1 < cmd.MinArguments() {
		usage(out, "Missing argument(s): expect=%d, actual=%d", cmd.MinArguments(), len(flag.Args())-1)
		os.Exit(int(syscall.EINVAL))
	}
	err = cmd.Run(flag.Args()[1:])
	if err != nil {
		log.Printf("Command '%s' failed: %s", cmdname, err.Error())
	}
}

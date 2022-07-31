package main

import "io"

type Command interface {
	Usage(w io.Writer, progname, cmdname string)
	MinArguments() int
	Run(args []string) error
}

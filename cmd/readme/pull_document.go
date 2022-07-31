package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type PullDocument struct {
	*RemoteCommand
}

func (c *PullDocument) Usage(w io.Writer, progname, cmdname string) {
	fmt.Fprintf(w, "%s %s <slug-like>\n\n", progname, cmdname)
	fmt.Fprintf(w, "Examples:\n\n")
	fmt.Fprintf(w, "%s %s quick-start\n", progname, cmdname)
	fmt.Fprintf(w, "%s %s https://foobar.com/docs/quick-start\n", progname, cmdname)
}

func (c *PullDocument) MinArguments() int {
	return 1
}

func (c *PullDocument) Run(args []string) error {
	slug := args[0]
	slug = filepath.Base(slug)
	slug = strings.TrimSuffix(slug, filepath.Ext(slug))
	return c.run(slug)
}

func (c *PullDocument) run(slug string) error {
	remote, err := c.client.Doc(slug)
	if err != nil {
		return err
	}
	cat, err := c.client.CategoryByID(remote.Category)
	if err != nil {
		return err
	}
	doc := RemoteDoc(cat.Slug, remote)
	old, err := LocalDoc(c.docRoot, slug)
	if err == nil {
		diff := c.diffDoc(slug, old, doc)
		if diff {
			cont, err := c.yesOrNo("Are you sure to pull '%s' and overwrite local changes?", slug)
			if err != nil {
				return err
			}
			if !cont {
				c.printf("Doc '%s' is not pulled", slug)
				return nil
			}
		} else {
			c.printf("Doc '%s' is alignmed with the remote one", slug)
			return nil
		}
	}
	err = doc.Save(c.docRoot, slug)
	if err != nil {
		return err
	}
	return nil
}

func (c *RemoteCommand) diffDoc(slug string, old, new *doc) bool {
	c.printf("Checking '%s' for difference...", slug)
	diff := false
	if old.Category != new.Category {
		c.printf("Category: %s => %s", old.Category, new.Category)
		diff = true
	}
	if old.Title != new.Title {
		c.printf("Title: %s => %s", old.Title, new.Title)
		diff = true
	}
	if old.Excerpt != new.Excerpt {
		c.printf("Excerpt: %s => %s", old.Excerpt, new.Excerpt)
		diff = true
	}
	if old.Hidden != new.Hidden {
		c.printf("Hidden: %v => %v", old.Hidden, new.Hidden)
		diff = true
	}
	if old.Body != new.Body {
		c.printf("Body:")
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(old.Body, new.Body, false)
		fmt.Println(dmp.DiffPrettyText(diffs))
		diff = true
	}
	return diff
}

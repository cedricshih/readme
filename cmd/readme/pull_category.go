package main

import (
	"fmt"

	"github.com/cedricshih/readme/api/readme"
)

type PullCategory struct {
	*RemoteCommand
}

func (c *PullCategory) MinArguments() int {
	return 0
}

func (c *PullCategory) Run(args []string) error {
	meta, err := c.metadata()
	if err != nil {
		return err
	}
	var cat *readme.Category
	var doc *readme.Doc
	docSlug := ""
	if len(args) > 0 {
		docSlug = args[0]
		doc, err = c.client.Doc(docSlug)
		if err != nil {
			return err
		}
		cats, err := c.client.Categories()
		if err != nil {
			return err
		}
		for _, cc := range cats {
			docs, err := c.client.Docs(cc.Slug)
			if err != nil {
				return err
			}
			for _, d := range docs {
				if d.Slug == docSlug {
					cat = cc
					break
				}
			}
			if cat != nil {
				break
			}
		}
		if cat == nil {
			return fmt.Errorf("unable to find the category of doc: '%s'", docSlug)
		}
	} else {
		var err error
		catSlug, err := c.chooseCategory(false)
		if err != nil {
			return err
		}
		if catSlug == "" {
			return nil
		}
		cat, err = c.client.Category(catSlug)
		if err != nil {
			return err
		}
		docSlug, err = c.chooseDoc(catSlug, true)
		if err != nil {
			return err
		}
		if docSlug == "" {
			return nil
		}
		if docSlug == selectionAll {
			changed, err := c.pullCategory(meta, cat)
			if err != nil {
				return err
			}
			if changed {
				return c.writeMetadata(meta)
			}
			return nil
		}
		doc, err = c.client.Doc(docSlug)
		if err != nil {
			return err
		}
	}
	changed, err := c.pullDoc(meta, cat, doc)
	if err != nil {
		return err
	}
	if changed {
		return c.writeMetadata(meta)
	}
	return nil
}

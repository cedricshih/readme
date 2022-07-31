package main

import (
	"fmt"
	"io/ioutil"

	"github.com/cedricshih/readme/api/readme"
)

type PushDocument struct {
	*RemoteCommand
}

func (c *PushDocument) MinArguments() int {
	return 0
}

func (c *PushDocument) Run(args []string) error {
	doc := ""
	if len(args) > 0 {
		doc = args[0]
	} else {
		category, err := c.chooseCategory(false)
		if err != nil {
			return err
		}
		if category == "" {
			return nil
		}
		doc, err = c.chooseDoc(category, false)
		if err != nil {
			return err
		}
		if doc == "" {
			return nil
		}
	}
	meta, err := c.metadata()
	if err != nil {
		return err
	}
	cat, catMeta, docMeta := meta.Doc(doc)
	if docMeta == nil {
		c.printf("Doc '%s' not found in '%s', please create the doc on ReadMe dashboard and do a 'pull'.", doc, c.metadataFilePath())
		return nil
	}
	path := c.docFilePath(cat, doc)
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	old, err := c.client.Doc(doc)
	if err != nil {
		return err
	}
	new := &readme.Doc{
		Slug:    doc,
		Title:   docMeta.Title,
		Excerpt: docMeta.Excerpt,
		Hidden:  docMeta.Hidden,
		Body:    string(body),
	}
	diff := c.diff(old, new)
	if !diff {
		c.printf("Doc '%s' is unchanged", doc)
		return nil
	}
	cont, err := c.yesOrNo("Are you sure to push doc '%s' to remote?", doc)
	if err != nil {
		return err
	}
	if !cont {
		c.printf("Doc '%s' is not pushed", doc)
		return nil
	}
	c.printf("Pushing to ReadMe: %s", path)
	err = c.client.UpdateDoc(catMeta.ID, new)
	if err != nil {
		return err
	}
	u := fmt.Sprintf("%s/docs/%s", meta.BaseURL, doc)
	c.printf("Doc '%s' is pushed to: %s", doc, u)
	return nil
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cedricshih/readme/api/readme"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v2"
)

const (
	selectionAll = "(all)"
)

type RemoteCommand struct {
	output  io.Writer
	input   io.Reader
	client  *readme.Client
	docRoot string
	allYes  bool
}

func (t *RemoteCommand) printf(format string, args ...interface{}) {
	fmt.Fprintf(t.output, format+"\n", args...)
}

func (c *RemoteCommand) pullCategory(meta *Metadata, cat *readme.Category) (bool, error) {
	changed := false
	docs, err := c.client.Docs(cat.Slug)
	if err != nil {
		return false, err
	}
	for _, doc := range docs {
		doc, err = c.client.Doc(doc.Slug)
		if err != nil {
			return false, err
		}
		chg, err := c.pullDoc(meta, cat, doc)
		if err != nil {
			return false, err
		}
		changed = changed || chg
	}
	return changed, nil
}

func (c *RemoteCommand) pullDoc(meta *Metadata, cat *readme.Category, doc *readme.Doc) (bool, error) {
	_, _, exist := meta.Doc(doc.Slug)
	if exist != nil {
		path := c.docFilePath(cat.Slug, doc.Slug)
		body, err := ioutil.ReadFile(path)
		if err != nil {
			return false, err
		}
		old := &readme.Doc{
			Slug:    doc.Slug,
			Title:   exist.Title,
			Excerpt: exist.Excerpt,
			Hidden:  exist.Hidden,
			Body:    string(body),
		}
		diff := c.diff(old, doc)
		if !diff {
			c.printf("Doc '%s' is not changed", doc.Slug)
			return false, nil
		}
		cont, err := c.yesOrNo("Are you sure to pull '%s' and overwrite local changes?", doc.Slug)
		if err != nil {
			return false, err
		}
		if !cont {
			c.printf("Doc '%s' is not pulled", doc.Slug)
			return false, nil
		}
	}
	if meta.Categories[cat.Slug] == nil {
		meta.Categories[cat.Slug] = &Category{
			ID:   cat.ID,
			Docs: make(map[string]*Doc),
		}
	}
	meta.Categories[cat.Slug].Docs[doc.Slug] = &Doc{
		Title:   doc.Title,
		Excerpt: doc.Excerpt,
		Hidden:  doc.Hidden,
	}
	path := c.docFilePath(cat.Slug, doc.Slug)
	err := os.MkdirAll(c.categoryPath(cat.Slug), os.ModePerm)
	if err != nil {
		return false, err
	}
	c.printf("Writing doc: %s", path)
	err = ioutil.WriteFile(path, []byte(doc.Body), os.ModePerm)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *RemoteCommand) metadata() (*Metadata, error) {
	prj, err := c.client.Project()
	if err != nil {
		return nil, err
	}
	meta := &Metadata{}
	path := c.metadataFilePath()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.printf("Creating new metadata...")
			meta.SubDomain = prj.SubDomain
			meta.Categories = make(map[string]*Category)
		} else {
			return nil, err
		}
	} else {
		err = yaml.Unmarshal(data, meta)
		if err != nil {
			return nil, err
		}
		if prj.SubDomain != meta.SubDomain {
			return nil, fmt.Errorf("the API key is from another project '%s' not '%s'", prj.SubDomain, meta.SubDomain)
		}
	}
	meta.BaseURL = prj.BaseUrl
	return meta, nil
}

func (c *RemoteCommand) writeMetadata(meta *Metadata) error {
	data, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}
	path := c.metadataFilePath()
	c.printf("Writing metadata: %s", path)
	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (c *RemoteCommand) metadataFilePath() string {
	return filepath.Join(c.docRoot, "metadata.yaml")
}

func (c *RemoteCommand) categoryPath(cat string) string {
	return filepath.Join(c.docRoot, cat)
}

func (c *RemoteCommand) docFilePath(cat, doc string) string {
	return filepath.Join(c.docRoot, cat, fmt.Sprintf("%s.md", doc))
}

func (c *RemoteCommand) diff(old, new *readme.Doc) bool {
	c.printf("Checking '%s' for difference...", old.Slug)
	diff := false
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

func (c *RemoteCommand) chooseCategory(all bool) (string, error) {
	res, err := c.client.Categories()
	if err != nil {
		return "", err
	}
	items := make([]string, 0)
	if all {
		items = append(items, selectionAll)
	}
	for _, c := range res {
		items = append(items, c.Slug)
	}
	_, category, err := c.choose(items, "Select category:")
	if err != nil {
		return "", err
	}
	return category, nil
}

func (c *RemoteCommand) chooseDoc(category string, all bool) (string, error) {
	res, err := c.client.Docs(category)
	if err != nil {
		return "", err
	}
	items := make([]string, 0)
	if all {
		items = append(items, selectionAll)
	}
	for _, d := range res {
		items = append(items, d.Slug)
	}
	_, doc, err := c.choose(items, "Select doc:")
	if err != nil {
		return "", err
	}
	return doc, nil
}

func (c *RemoteCommand) choose(items []string, format string, args ...interface{}) (int, string, error) {
	c.prompt(items, format, args...)
	return c.receiveSelection(items)
}

func (c *RemoteCommand) prompt(items []string, format string, args ...interface{}) {
	c.printf(format, args...)
	for i, it := range items {
		c.printf("%d:\t%s", i+1, it)
	}
}

func (c *RemoteCommand) receiveSelection(items []string) (int, string, error) {
	reader := bufio.NewReader(c.input)
	text, err := reader.ReadString('\n')
	if err != nil {
		return -1, "", err
	}
	text = strings.TrimSpace(text)
	num, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		return -1, "", nil
	}
	if num <= 0 || int(num) > len(items) {
		return -1, "", nil
	}
	return int(num) - 1, items[num-1], nil
}

func (c *RemoteCommand) yesOrNo(format string, args ...interface{}) (bool, error) {
	if c.allYes {
		return true, nil
	}
	c.printf(format+" (Y/n)", args...)
	reader := bufio.NewReader(c.input)
	b, err := reader.ReadByte()
	if err != nil {
		return false, err
	}
	switch b {
	case 'Y':
		return true, nil
	default:
		return false, nil
	}
}

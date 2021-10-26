package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"readme/api/readme"
	"strconv"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v2"
)

const (
	selectionAll = "(all)"
)

type tool struct {
	output  io.Writer
	input   io.Reader
	client  *readme.Client
	docRoot string
	allYes  bool
}

func (t *tool) printf(format string, args ...interface{}) {
	fmt.Fprintf(t.output, format+"\n", args...)
}

func (t *tool) categories() error {
	res, err := t.client.Categories()
	if err != nil {
		return err
	}
	if t.client.Output != nil {
		return nil
	}
	t.printf("Got %d categories:", len(res))
	for _, c := range res {
		t.printf("- %s : %s", c.Slug, c.Title)
	}
	return nil
}

func (t *tool) docs(category string) error {
	if category == "" {
		var err error
		category, err = t.chooseCategory(false)
		if err != nil {
			return err
		}
		if category == "" {
			return nil
		}
	}
	docs, err := t.client.Docs(category)
	if err != nil {
		return err
	}
	if t.client.Output != nil {
		return nil
	}
	t.printf("Got %d docs in '%s':", len(docs), category)
	for _, d := range docs {
		if d.Hidden {
			t.printf("- %s : %s (hidden)", d.Slug, d.Title)
		} else {
			t.printf("- %s : %s", d.Slug, d.Title)
		}
	}
	return nil
}

func (t *tool) doc(doc string) error {
	if doc == "" {
		category, err := t.chooseCategory(false)
		if err != nil {
			return err
		}
		if category == "" {
			return nil
		}
		doc, err = t.chooseDoc(category, false)
		if err != nil {
			return err
		}
		if doc == "" {
			return nil
		}
	}
	res, err := t.client.Doc(doc)
	if err != nil {
		return err
	}
	if t.client.Output != nil {
		return nil
	}
	if res.Hidden {
		t.printf("Title: %s (hidden)", res.Title)
	} else {
		t.printf("Title: %s", res.Title)
	}
	t.printf("Excerpt: %s", res.Excerpt)
	t.printf("Body:")
	t.printf("%s", res.Body)
	return nil
}

func (t *tool) pull(docSlug string) error {
	meta, err := t.metadata()
	if err != nil {
		return err
	}
	var cat *readme.Category
	var doc *readme.Doc
	if docSlug != "" {
		doc, err = t.client.Doc(docSlug)
		if err != nil {
			return err
		}
		cats, err := t.client.Categories()
		if err != nil {
			return err
		}
		for _, c := range cats {
			docs, err := t.client.Docs(c.Slug)
			if err != nil {
				return err
			}
			for _, d := range docs {
				if d.Slug == docSlug {
					cat = c
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
		catSlug, err := t.chooseCategory(false)
		if err != nil {
			return err
		}
		if catSlug == "" {
			return nil
		}
		cat, err = t.client.Category(catSlug)
		if err != nil {
			return err
		}
		docSlug, err = t.chooseDoc(catSlug, true)
		if err != nil {
			return err
		}
		if docSlug == "" {
			return nil
		}
		if docSlug == selectionAll {
			changed, err := t.pullCategory(meta, cat)
			if err != nil {
				return err
			}
			if changed {
				return t.writeMetadata(meta)
			}
			return nil
		}
		doc, err = t.client.Doc(docSlug)
		if err != nil {
			return err
		}
	}
	changed, err := t.pullDoc(meta, cat, doc)
	if err != nil {
		return err
	}
	if changed {
		return t.writeMetadata(meta)
	}
	return nil
}

func (t *tool) pullAll() error {
	meta, err := t.metadata()
	if err != nil {
		return err
	}
	cats, err := t.client.Categories()
	if err != nil {
		return err
	}
	changed := false
	for _, cat := range cats {
		chg, err := t.pullCategory(meta, cat)
		if err != nil {
			return err
		}
		changed = changed || chg
	}
	if changed {
		return t.writeMetadata(meta)
	}
	return nil
}

func (t *tool) pullCategory(meta *Metadata, cat *readme.Category) (bool, error) {
	changed := false
	docs, err := t.client.Docs(cat.Slug)
	if err != nil {
		return false, err
	}
	for _, doc := range docs {
		doc, err = t.client.Doc(doc.Slug)
		if err != nil {
			return false, err
		}
		chg, err := t.pullDoc(meta, cat, doc)
		if err != nil {
			return false, err
		}
		changed = changed || chg
	}
	return changed, nil
}

func (t *tool) pullDoc(meta *Metadata, cat *readme.Category, doc *readme.Doc) (bool, error) {
	_, _, exist := meta.Doc(doc.Slug)
	if exist != nil {
		path := t.docFilePath(cat.Slug, doc.Slug)
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
		diff := t.diff(old, doc)
		if !diff {
			t.printf("Doc '%s' is not changed", doc.Slug)
			return false, nil
		}
		cont, err := t.yesOrNo("Are you sure to pull '%s' and overwrite local changes?", doc.Slug)
		if err != nil {
			return false, err
		}
		if !cont {
			t.printf("Doc '%s' is not pulled", doc.Slug)
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
	path := t.docFilePath(cat.Slug, doc.Slug)
	err := os.MkdirAll(t.categoryPath(cat.Slug), os.ModePerm)
	if err != nil {
		return false, err
	}
	t.printf("Writing doc: %s", path)
	err = ioutil.WriteFile(path, []byte(doc.Body), os.ModePerm)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (t *tool) push(doc string) error {
	if doc == "" {
		category, err := t.chooseCategory(false)
		if err != nil {
			return err
		}
		if category == "" {
			return nil
		}
		doc, err = t.chooseDoc(category, false)
		if err != nil {
			return err
		}
		if doc == "" {
			return nil
		}
	}
	meta, err := t.metadata()
	if err != nil {
		return err
	}
	cat, catMeta, docMeta := meta.Doc(doc)
	if docMeta == nil {
		t.printf("Doc '%s' not found in '%s', please create the doc on ReadMe dashboard and do a 'pull'.", doc, t.metadataFilePath())
		return nil
	}
	path := t.docFilePath(cat, doc)
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	old, err := t.client.Doc(doc)
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
	diff := t.diff(old, new)
	if !diff {
		t.printf("Doc '%s' is unchanged", doc)
		return nil
	}
	cont, err := t.yesOrNo("Are you sure to push doc '%s' to remote?", doc)
	if err != nil {
		return err
	}
	if !cont {
		t.printf("Doc '%s' is not pushed", doc)
		return nil
	}
	t.printf("Pushing to ReadMe: %s", path)
	err = t.client.UpdateDoc(catMeta.ID, new)
	if err != nil {
		return err
	}
	u := fmt.Sprintf("%s/docs/%s", meta.BaseURL, doc)
	t.printf("Doc '%s' is pushed to: %s", doc, u)
	return nil
}

func (t *tool) metadata() (*Metadata, error) {
	prj, err := t.client.Project()
	if err != nil {
		return nil, err
	}
	meta := &Metadata{}
	path := t.metadataFilePath()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.printf("Creating new metadata...")
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

func (t *tool) writeMetadata(meta *Metadata) error {
	data, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}
	path := t.metadataFilePath()
	t.printf("Writing metadata: %s", path)
	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (t *tool) metadataFilePath() string {
	return filepath.Join(t.docRoot, "metadata.yaml")
}

func (t *tool) categoryPath(cat string) string {
	return filepath.Join(t.docRoot, cat)
}

func (t *tool) docFilePath(cat, doc string) string {
	return filepath.Join(t.docRoot, cat, fmt.Sprintf("%s.md", doc))
}

func (t *tool) diff(old, new *readme.Doc) bool {
	t.printf("Checking '%s' for difference...", old.Slug)
	diff := false
	if old.Title != new.Title {
		t.printf("Title: %s => %s", old.Title, new.Title)
		diff = true
	}
	if old.Excerpt != new.Excerpt {
		t.printf("Excerpt: %s => %s", old.Excerpt, new.Excerpt)
		diff = true
	}
	if old.Hidden != new.Hidden {
		t.printf("Hidden: %v => %v", old.Hidden, new.Hidden)
		diff = true
	}
	if old.Body != new.Body {
		t.printf("Body:")
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(old.Body, new.Body, false)
		fmt.Println(dmp.DiffPrettyText(diffs))
		diff = true
	}
	return diff
}

func (t *tool) chooseCategory(all bool) (string, error) {
	res, err := t.client.Categories()
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
	_, category, err := t.choose(items, "Select category:")
	if err != nil {
		return "", err
	}
	return category, nil
}

func (t *tool) chooseDoc(category string, all bool) (string, error) {
	res, err := t.client.Docs(category)
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
	_, doc, err := t.choose(items, "Select doc:")
	if err != nil {
		return "", err
	}
	return doc, nil
}

func (t *tool) choose(items []string, format string, args ...interface{}) (int, string, error) {
	t.prompt(items, format, args...)
	return t.receiveSelection(items)
}

func (t *tool) prompt(items []string, format string, args ...interface{}) {
	t.printf(format, args...)
	for i, it := range items {
		t.printf("%d:\t%s", i+1, it)
	}
}

func (t *tool) receiveSelection(items []string) (int, string, error) {
	reader := bufio.NewReader(t.input)
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

func (t *tool) yesOrNo(format string, args ...interface{}) (bool, error) {
	if t.allYes {
		return true, nil
	}
	t.printf(format+" (Y/n)", args...)
	reader := bufio.NewReader(t.input)
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

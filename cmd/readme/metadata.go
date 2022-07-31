package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/cedricshih/readme/api/readme"
	"gopkg.in/yaml.v2"
)

type Metadata struct {
	SubDomain  string
	BaseURL    string
	Categories map[string]*Category
}

func (m *Metadata) Doc(slug string) (string, *Category, *Doc) {
	for catKey, cat := range m.Categories {
		for docKey, doc := range cat.Docs {
			if docKey == slug {
				return catKey, cat, doc
			}
		}
	}
	return "", nil, nil
}

type Category struct {
	ID   string
	Docs map[string]*Doc
}

type Doc struct {
	Title   string
	Excerpt string
	Hidden  bool
}

type doc struct {
	Category string
	Title    string
	Excerpt  string
	Hidden   bool
	Body     string `yaml:"-"`
}

func RemoteDoc(cat string, remote *readme.Doc) *doc {
	return &doc{
		Category: cat,
		Title:    remote.Title,
		Excerpt:  remote.Excerpt,
		Hidden:   remote.Hidden,
		Body:     remote.Body,
	}
}

func LocalDoc(dir, slug string) (*doc, error) {
	bodyFilename := filepath.Join(dir, fmt.Sprintf("%s.md", slug))
	metaFilename := filepath.Join(dir, fmt.Sprintf("%s.yaml", slug))
	body, err := os.Open(bodyFilename)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	meta, err := os.Open(metaFilename)
	if err != nil {
		return nil, err
	}
	defer meta.Close()
	doc := &doc{}
	err = yaml.NewDecoder(meta).Decode(doc)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	doc.Body = string(data)
	return doc, nil
}

func (d *doc) Save(dir, slug string) error {
	bodyFilename := filepath.Join(dir, fmt.Sprintf("%s.md", slug))
	metaFilename := filepath.Join(dir, fmt.Sprintf("%s.yaml", slug))
	meta, err := os.Create(metaFilename)
	if err != nil {
		return err
	}
	log.Printf("Writing meta: %s", metaFilename)
	err = yaml.NewEncoder(meta).Encode(d)
	if err != nil {
		return err
	}
	log.Printf("Writing body: %s", bodyFilename)
	err = ioutil.WriteFile(bodyFilename, []byte(d.Body), 0644)
	if err != nil {
		return err
	}
	return nil
}

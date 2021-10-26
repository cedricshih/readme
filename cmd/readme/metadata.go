package main

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

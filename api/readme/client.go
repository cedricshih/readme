package readme

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/TylerBrock/colorjson"
)

type Client struct {
	*http.Client
	Endpoint      string
	APIKey        string
	Output        io.Writer
	categoryCache map[string]*Category
}

func NewClient(APIKey string) *Client {
	return &Client{
		Client:        http.DefaultClient,
		Endpoint:      "https://dash.readme.com/api/v1/",
		APIKey:        APIKey,
		categoryCache: map[string]*Category{},
	}
}

func (c *Client) Project() (*Project, error) {
	res := &Project{}
	err := c.request("GET", "", nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) Categories() ([]*Category, error) {
	res := make([]*Category, 0)
	if len(c.categoryCache) > 0 {
		for _, v := range c.categoryCache {
			res = append(res, v)
		}
		return res, nil
	}
	page := 1
	for {
		cats, err := c.categories(page)
		if err != nil {
			return nil, err
		}
		for _, cat := range cats {
			c.categoryCache[cat.ID] = cat
		}
		res = append(res, cats...)
		if len(res) < 100 {
			break
		}
		page++
	}
	return res, nil
}

func (c *Client) categories(page int) ([]*Category, error) {
	res := make([]*Category, 0)
	err := c.request("GET", fmt.Sprintf("categories?perPage=100&page=%d", page), nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) Category(category string) (*Category, error) {
	res := &Category{}
	err := c.request("GET", fmt.Sprintf("categories/%s", category), nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) CategoryByID(id string) (*Category, error) {
	if len(c.categoryCache) <= 0 {
		_, err := c.Categories()
		if err != nil {
			return nil, err
		}
	}
	cat := c.categoryCache[id]
	if cat == nil {
		return nil, fmt.Errorf("no such category: %s", id)
	}
	return cat, nil
}

func (c *Client) Docs(category string) ([]*Doc, error) {
	res := make([]*Doc, 0)
	err := c.request("GET", fmt.Sprintf("categories/%s/docs", category), nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) Doc(doc string) (*Doc, error) {
	res := &Doc{}
	err := c.request("GET", fmt.Sprintf("docs/%s", doc), nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) UpdateDoc(cat string, doc *Doc) error {
	req := &struct {
		Title    string `json:"title"`
		Body     string `json:"body,omitempty"`
		Category string `json:"category"`
		Hiddle   bool   `json:"hidden"`
	}{
		Title:    doc.Title,
		Body:     doc.Body,
		Category: cat,
		Hiddle:   doc.Hidden,
	}
	return c.request("PUT", fmt.Sprintf("docs/%s", doc.Slug), req, nil)
}

func (c *Client) request(method, uri string, reqJson interface{}, resJson interface{}) error {
	var body io.Reader
	if reqJson != nil {
		data, err := json.Marshal(reqJson)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(data)
		if c.Output != nil {
			err = c.prettyPrint(data)
			if err != nil {
				return err
			}
		}
	}
	req, err := http.NewRequest(method, c.Endpoint+uri, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.RawStdEncoding.EncodeToString([]byte(c.APIKey+":")))
	log.Printf("Making request: %s %s", req.Method, req.URL.String())
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if c.Output != nil {
		err = c.prettyPrint(data)
		if err != nil {
			return err
		}
		// io.Copy(c.Output, bytes.NewBuffer(data))
	}
	if res.StatusCode != 200 {
		readmeErr := &Error{}
		err = json.Unmarshal(data, readmeErr)
		if err != nil {
			return errors.New(res.Status)
		} else {
			return readmeErr
		}
	}
	if resJson != nil {
		err = json.Unmarshal(data, resJson)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) prettyPrint(data []byte) error {
	var obj interface{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}
	f := colorjson.NewFormatter()
	f.Indent = 2
	out, err := f.Marshal(obj)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.Output, "%s\n", out)
	return nil
}

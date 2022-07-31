package main

type GetDocument struct {
	*RemoteCommand
}

func (c *GetDocument) MinArguments() int {
	return 0
}

func (c *GetDocument) Run(args []string) error {
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
	res, err := c.client.Doc(doc)
	if err != nil {
		return err
	}
	if c.client.Output != nil {
		return nil
	}
	if res.Hidden {
		c.printf("Title: %s (hidden)", res.Title)
	} else {
		c.printf("Title: %s", res.Title)
	}
	c.printf("Excerpt: %s", res.Excerpt)
	c.printf("Body:")
	c.printf("%s", res.Body)
	return nil
}

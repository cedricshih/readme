package main

type ListDocuments struct {
	*RemoteCommand
}

func (c *ListDocuments) MinArguments() int {
	return 1
}

func (c *ListDocuments) Run(args []string) error {
	category := args[0]
	if category == "" {
		var err error
		category, err = c.chooseCategory(false)
		if err != nil {
			return err
		}
		if category == "" {
			return nil
		}
	}
	docs, err := c.client.Docs(category)
	if err != nil {
		return err
	}
	if c.client.Output != nil {
		return nil
	}
	c.printf("Got %d docs in '%s':", len(docs), category)
	for _, d := range docs {
		if d.Hidden {
			c.printf("- %s : %s (hidden)", d.Slug, d.Title)
		} else {
			c.printf("- %s : %s", d.Slug, d.Title)
		}
	}
	return nil
}

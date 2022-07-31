package main

type ListCategories struct {
	*RemoteCommand
}

func (c *ListCategories) MinArguments() int {
	return 0
}

func (c *ListCategories) Run(args []string) error {
	res, err := c.client.Categories()
	if err != nil {
		return err
	}
	if c.client.Output != nil {
		return nil
	}
	c.printf("Got %d categories:", len(res))
	for _, cat := range res {
		c.printf("- %s : %s", cat.Slug, cat.Title)
	}
	return nil
}

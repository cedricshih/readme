package main

type GetCategory struct {
	*RemoteCommand
}

func (c *GetCategory) MinArguments() int {
	return 1
}

func (c *GetCategory) Run(args []string) error {
	cat := args[0]
	res, err := c.client.Category(cat)
	if err != nil {
		return err
	}
	if c.client.Output != nil {
		return nil
	}
	c.printf("ID:    %s", res.ID)
	c.printf("Slug:  %s", res.Slug)
	c.printf("Title: %s", res.Title)
	return nil
}

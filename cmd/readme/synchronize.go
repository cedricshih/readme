package main

type Synchronize struct {
	*RemoteCommand
}

func (c *Synchronize) MinArguments() int {
	return 0
}

func (c *Synchronize) Run(args []string) error {
	meta, err := c.metadata()
	if err != nil {
		return err
	}
	cats, err := c.client.Categories()
	if err != nil {
		return err
	}
	changed := false
	for _, cat := range cats {
		chg, err := c.pullCategory(meta, cat)
		if err != nil {
			return err
		}
		changed = changed || chg
	}
	if changed {
		return c.writeMetadata(meta)
	}
	return nil
}

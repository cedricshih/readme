package readme

type Doc struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Excerpt string `json:"excerpt"`
	Body    string `json:"body"`
	Hidden  bool   `json:"hidden"`
}

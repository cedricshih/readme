package readme

type Doc struct {
	Slug     string `json:"slug"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Excerpt  string `json:"excerpt"`
	Body     string `json:"body" yaml:"-"`
	Hidden   bool   `json:"hidden"`
}

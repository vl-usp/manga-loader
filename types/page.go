package types

type Page struct {
	ID    int    `json:"id"`
	Image string `json:"image"`
	Slug  int    `json:"slug"`
	URL   string `json:"url"`
}

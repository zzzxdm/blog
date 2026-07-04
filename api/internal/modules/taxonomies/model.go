package taxonomies

type Category struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
	PostCount   int    `json:"postCount"`
}

type Tag struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	PostCount int    `json:"postCount"`
}

type SaveCategoryRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

type SaveTagRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

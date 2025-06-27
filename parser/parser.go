package parser

// this is where the links are being stored
type Token struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Info struct {
	Title       string   `json:"title"`
	StatusCode  int      `json:"statusCode"`
	Description string   `json:"description"`
	Paragraphs  []string `json:"paragraphs"`
	Links       []Token  `json:"links"`
}

type Parser interface {
	IsSupportedExtension(extension string) bool
	Parse(content string) (Info, error)
}

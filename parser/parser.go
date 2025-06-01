package parser

// this is where the links are being stored
type Token struct {
	Name  string
	Value string
}

type Info struct {
	Title       string
	Description string
	Paragraphs  []string
	Links       []Token
}

type Parser interface {
	IsSupportedExtension(extension string) bool
	Parse(content string) (Info, error)
}

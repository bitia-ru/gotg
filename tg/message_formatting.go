package tg

import (
	"github.com/gotd/td/telegram/message/styling"
)

// MessageChunk represents a piece of a formatted message
type MessageChunk interface {
	// ToStyledTextOptions converts the chunk to gotd StyledTextOption slice
	ToStyledTextOptions() []styling.StyledTextOption
}

// TextChunk represents a plain text chunk
type TextChunk struct {
	Text string
}

// NewTextChunk creates a new plain text chunk
func NewTextChunk(text string) *TextChunk {
	return &TextChunk{Text: text}
}

func (t *TextChunk) ToStyledTextOptions() []styling.StyledTextOption {
	return []styling.StyledTextOption{styling.Plain(t.Text)}
}

// StyledTextChunk represents a text chunk with styling
type StyledTextChunk struct {
	Text  string
	Style TextStyle
}

// TextStyle represents the style of text
type TextStyle int

const (
	StylePlain TextStyle = iota
	StyleBold
	StyleItalic
	StyleCode
	StylePre
	StyleStrike
	StyleUnderline
	StyleSpoiler
)

// NewStyledTextChunk creates a new styled text chunk
func NewStyledTextChunk(text string, style TextStyle) *StyledTextChunk {
	return &StyledTextChunk{Text: text, Style: style}
}

func (s *StyledTextChunk) ToStyledTextOptions() []styling.StyledTextOption {
	switch s.Style {
	case StyleBold:
		return []styling.StyledTextOption{styling.Bold(s.Text)}
	case StyleItalic:
		return []styling.StyledTextOption{styling.Italic(s.Text)}
	case StyleCode:
		return []styling.StyledTextOption{styling.Code(s.Text)}
	case StylePre:
		return []styling.StyledTextOption{styling.Pre(s.Text, "")}
	case StyleStrike:
		return []styling.StyledTextOption{styling.Strike(s.Text)}
	case StyleUnderline:
		return []styling.StyledTextOption{styling.Underline(s.Text)}
	case StyleSpoiler:
		return []styling.StyledTextOption{styling.Spoiler(s.Text)}
	default:
		return []styling.StyledTextOption{styling.Plain(s.Text)}
	}
}

// ContainerChunk represents a chunk that contains other chunks
type ContainerChunk struct {
	Chunks []MessageChunk
}

// NewContainerChunk creates a new container chunk
func NewContainerChunk(chunks ...MessageChunk) *ContainerChunk {
	return &ContainerChunk{Chunks: chunks}
}

func (c *ContainerChunk) ToStyledTextOptions() []styling.StyledTextOption {
	var options []styling.StyledTextOption
	for _, chunk := range c.Chunks {
		options = append(options, chunk.ToStyledTextOptions()...)
	}
	return options
}

// Add adds a chunk to the container
func (c *ContainerChunk) Add(chunk MessageChunk) {
	c.Chunks = append(c.Chunks, chunk)
}

// LinkChunk represents a text chunk with URL
type LinkChunk struct {
	Text string
	URL  string
}

// NewLinkChunk creates a new link chunk
func NewLinkChunk(text, url string) *LinkChunk {
	return &LinkChunk{Text: text, URL: url}
}

func (l *LinkChunk) ToStyledTextOptions() []styling.StyledTextOption {
	return []styling.StyledTextOption{styling.TextURL(l.Text, l.URL)}
}

// PreCodeChunk represents a code block with language
type PreCodeChunk struct {
	Code     string
	Language string
}

// NewPreCodeChunk creates a new pre-formatted code chunk
func NewPreCodeChunk(code, language string) *PreCodeChunk {
	return &PreCodeChunk{Code: code, Language: language}
}

func (p *PreCodeChunk) ToStyledTextOptions() []styling.StyledTextOption {
	return []styling.StyledTextOption{styling.Pre(p.Code, p.Language)}
}

// Helper functions for common formatting patterns

// Bold creates a bold text chunk
func Bold(text string) *StyledTextChunk {
	return NewStyledTextChunk(text, StyleBold)
}

// Italic creates an italic text chunk
func Italic(text string) *StyledTextChunk {
	return NewStyledTextChunk(text, StyleItalic)
}

// Code creates an inline code chunk
func Code(text string) *StyledTextChunk {
	return NewStyledTextChunk(text, StyleCode)
}

// Strike creates a strikethrough text chunk
func Strike(text string) *StyledTextChunk {
	return NewStyledTextChunk(text, StyleStrike)
}

// Spoiler creates a spoiler text chunk
func Spoiler(text string) *StyledTextChunk {
	return NewStyledTextChunk(text, StyleSpoiler)
}

// Underline creates an underlined text chunk
func Underline(text string) *StyledTextChunk {
	return NewStyledTextChunk(text, StyleUnderline)
}

// Link creates a text link chunk
func Link(text, url string) *LinkChunk {
	return NewLinkChunk(text, url)
}

// PreCode creates a pre-formatted code block
func PreCode(code, language string) *PreCodeChunk {
	return NewPreCodeChunk(code, language)
}

// Text creates a plain text chunk
func Text(text string) *TextChunk {
	return NewTextChunk(text)
}

// Container creates a container chunk with the given chunks
func Container(chunks ...MessageChunk) *ContainerChunk {
	return NewContainerChunk(chunks...)
}
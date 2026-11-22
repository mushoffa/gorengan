package csv

import (
	"time"
)

type config struct {
	delimiter  rune
	nilValue   string
	timeFormat string
	noHeader   bool
	crlf       bool
}

func defaultConfig() config {
	return config{
		delimiter:  ',',
		timeFormat: time.RFC3339,
	}
}

func applyOptions(opts Options) config {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

type Options []Option
type Option func(*config)

// WithDelimiter sets the field separator (default ',').
func WithDelimiter(r rune) Option { return func(c *config) { c.delimiter = r } }

// WithoutHeader skips the header row.
func WithoutHeader() Option { return func(c *config) { c.noHeader = true } }

// WithCRLF uses \r\n line endings.
func WithCRLF() Option { return func(c *config) { c.crlf = true } }

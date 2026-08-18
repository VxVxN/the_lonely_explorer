package simpleui

import (
	"bytes"

	text "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// LoadFont parses TTF/OTF font bytes and returns a text/v2 Face at the given
// pixel size. Callers typically load ttfBytes via go:embed and call this
// once per size they need.
func LoadFont(ttfBytes []byte, size float64) (text.Face, error) {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(ttfBytes))
	if err != nil {
		return nil, err
	}
	return &text.GoTextFace{
		Source: source,
		Size:   size,
	}, nil
}

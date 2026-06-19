// Package images produces deterministic placeholder WebP assets for the
// MONTI demo seed. Each asset is a solid-color rectangle with the item or
// category name centered in white using basicfont.Face7x13. WebP encoding
// goes through github.com/HugoSmits86/nativewebp — pure Go, no cgo.
package images

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/HugoSmits86/nativewebp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Encoder returns the encoder identifier used in startup log lines so that
// downstream tests (TEST-0008 evidence) can record which path produced the
// uploaded objects.
const Encoder = "HugoSmits86/nativewebp v1.3.0 (pure Go)"

// Generate renders a solid-color WebP at the given dimensions with the label
// centered horizontally and vertically. RGB triple is the fill color.
func Generate(w, h int, label string, rgb [3]uint8) ([]byte, error) {
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("images.Generate: invalid dimensions %dx%d", w, h)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fill := color.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 0xff}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: fill}, image.Point{}, draw.Src)

	if label != "" {
		face := basicfont.Face7x13
		// basicfont glyph width is 7px, height 13px.
		textWidth := len(label) * 7
		x := (w - textWidth) / 2
		y := (h+13)/2 - 2
		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(color.White),
			Face: face,
			Dot:  fixed.P(x, y),
		}
		d.DrawString(label)
	}

	var buf bytes.Buffer
	if err := nativewebp.Encode(&buf, img, nil); err != nil {
		return nil, fmt.Errorf("images.Generate: webp encode: %w", err)
	}
	return buf.Bytes(), nil
}

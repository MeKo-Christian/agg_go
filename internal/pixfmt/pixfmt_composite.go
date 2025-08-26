package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

// PixFmtCompositeRGBA represents an RGBA pixel format with composite blending
type PixFmtCompositeRGBA[CS any, O any] struct {
	rbuf    *buffer.RenderingBufferU8
	blender blender.CompositeBlender[CS, O]
}

// NewPixFmtCompositeRGBA creates a new composite RGBA pixel format
func NewPixFmtCompositeRGBA[CS any, O any](rbuf *buffer.RenderingBufferU8, op blender.CompOp) *PixFmtCompositeRGBA[CS, O] {
	return &PixFmtCompositeRGBA[CS, O]{
		rbuf:    rbuf,
		blender: blender.NewCompositeBlender[CS, O](op),
	}
}

// Width returns the buffer width
func (pf *PixFmtCompositeRGBA[CS, O]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtCompositeRGBA[CS, O]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (4 for RGBA)
func (pf *PixFmtCompositeRGBA[CS, O]) PixWidth() int {
	return 4
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtCompositeRGBA[CS, O]) GetPixel(x, y int) color.RGBA8[CS] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA8[CS]{}
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return color.RGBA8[CS]{}
	}

	order := blender.GetColorOrder[O]()
	return color.RGBA8[CS]{
		R: row[pixelOffset+int(order.R)],
		G: row[pixelOffset+int(order.G)],
		B: row[pixelOffset+int(order.B)],
		A: row[pixelOffset+int(order.A)],
	}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtCompositeRGBA[CS, O]) CopyPixel(x, y int, c color.RGBA8[CS]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return
	}

	order := blender.GetColorOrder[O]()
	row[pixelOffset+int(order.R)] = c.R
	row[pixelOffset+int(order.G)] = c.G
	row[pixelOffset+int(order.B)] = c.B
	row[pixelOffset+int(order.A)] = c.A
}

// BlendPixel blends a pixel using composite blending
func (pf *PixFmtCompositeRGBA[CS, O]) BlendPixel(x, y int, c color.RGBA8[CS], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return
	}

	// Use the composite blender to perform the blending operation
	pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, cover)
}

// BlendHline blends a horizontal line of pixels (interface method)
func (pf *PixFmtCompositeRGBA[CS, O]) BlendHline(x, y, length int, c color.RGBA8[CS], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() {
		return
	}

	startX := max(0, x)
	endX := min(x+length, pf.Width())

	if startX >= endX {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)

	for i := startX; i < endX; i++ {
		pixelOffset := i * 4
		if pixelOffset+3 < len(row) {
			pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, cover)
		}
	}
}

// BlendVline blends a vertical line of pixels (interface method)
func (pf *PixFmtCompositeRGBA[CS, O]) BlendVline(x, y, length int, c color.RGBA8[CS], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() {
		return
	}

	startY := max(0, y)
	endY := min(y+length, pf.Height())

	if startY >= endY {
		return
	}

	pixelOffset := x * 4

	for i := startY; i < endY; i++ {
		row := buffer.RowU8(pf.rbuf, i)
		if pixelOffset+3 < len(row) {
			pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, cover)
		}
	}
}

// BlendSolidHspan blends a horizontal span with solid color (interface method)
func (pf *PixFmtCompositeRGBA[CS, O]) BlendSolidHspan(x, y, length int, c color.RGBA8[CS], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() {
		return
	}

	startX := max(0, x)
	endX := min(x+length, pf.Width())

	if startX >= endX {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)

	for i := startX; i < endX; i++ {
		pixelOffset := i * 4
		coverIndex := i - x
		if pixelOffset+3 < len(row) && coverIndex < len(covers) {
			pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, covers[coverIndex])
		}
	}
}

// SetCompOp changes the composite operation
func (pf *PixFmtCompositeRGBA[CS, O]) SetCompOp(op blender.CompOp) {
	pf.blender = blender.NewCompositeBlender[CS, O](op)
}

// GetCompOp returns the current composite operation
func (pf *PixFmtCompositeRGBA[CS, O]) GetCompOp() blender.CompOp {
	return pf.blender.GetOp()
}

// Pixel returns the pixel at the given coordinates (alias for GetPixel)
func (pf *PixFmtCompositeRGBA[CS, O]) Pixel(x, y int) color.RGBA8[CS] {
	return pf.GetPixel(x, y)
}

// CopyHline copies a horizontal line of pixels without blending
func (pf *PixFmtCompositeRGBA[CS, O]) CopyHline(x, y, length int, c color.RGBA8[CS]) {
	for i := 0; i < length; i++ {
		pf.CopyPixel(x+i, y, c)
	}
}

// CopyVline copies a vertical line of pixels without blending
func (pf *PixFmtCompositeRGBA[CS, O]) CopyVline(x, y, length int, c color.RGBA8[CS]) {
	for i := 0; i < length; i++ {
		pf.CopyPixel(x, y+i, c)
	}
}

// BlendVline blends a vertical line of pixels (already implemented above)
// CopyBar copies a rectangular area with solid color
func (pf *PixFmtCompositeRGBA[CS, O]) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[CS]) {
	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2-x1+1, c)
	}
}

// BlendBar blends a rectangular area with solid color
func (pf *PixFmtCompositeRGBA[CS, O]) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[CS], cover basics.Int8u) {
	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2-x1+1, c, cover)
	}
}

// BlendSolidVspan blends a vertical span with solid color
func (pf *PixFmtCompositeRGBA[CS, O]) BlendSolidVspan(x, y, length int, c color.RGBA8[CS], covers []basics.Int8u) {
	for i := 0; i < length && i < len(covers); i++ {
		pf.BlendPixel(x, y+i, c, covers[i])
	}
}

// CopyColorHspan copies a horizontal span with varying colors
func (pf *PixFmtCompositeRGBA[CS, O]) CopyColorHspan(x, y, length int, colors []color.RGBA8[CS]) {
	for i := 0; i < length && i < len(colors); i++ {
		pf.CopyPixel(x+i, y, colors[i])
	}
}

// BlendColorHspan blends a horizontal span with varying colors
func (pf *PixFmtCompositeRGBA[CS, O]) BlendColorHspan(x, y, length int, colors []color.RGBA8[CS], covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		actualCover := cover
		if i < len(covers) {
			actualCover = covers[i]
		}
		pf.BlendPixel(x+i, y, colors[i], actualCover)
	}
}

// CopyColorVspan copies a vertical span with varying colors
func (pf *PixFmtCompositeRGBA[CS, O]) CopyColorVspan(x, y, length int, colors []color.RGBA8[CS]) {
	for i := 0; i < length && i < len(colors); i++ {
		pf.CopyPixel(x, y+i, colors[i])
	}
}

// BlendColorVspan blends a vertical span with varying colors
func (pf *PixFmtCompositeRGBA[CS, O]) BlendColorVspan(x, y, length int, colors []color.RGBA8[CS], covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		actualCover := cover
		if i < len(covers) {
			actualCover = covers[i]
		}
		pf.BlendPixel(x, y+i, colors[i], actualCover)
	}
}

// Clear fills the entire buffer with the specified color
func (pf *PixFmtCompositeRGBA[CS, O]) Clear(c color.RGBA8[CS]) {
	// Fill entire buffer
	width := pf.Width()
	height := pf.Height()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pf.CopyPixel(x, y, c)
		}
	}
}

// Fill fills the entire buffer with the specified color (alias for Clear)
func (pf *PixFmtCompositeRGBA[CS, O]) Fill(c color.RGBA8[CS]) {
	pf.Clear(c)
}

// Note: max and min functions are already defined in pixfmt_rgba64.go, so we don't redefine them here

// Concrete composite pixel format types for convenience
type (
	PixFmtCompositeRGBA32     = PixFmtCompositeRGBA[color.Linear, blender.RGBAOrder]
	PixFmtCompositeARGB32     = PixFmtCompositeRGBA[color.Linear, blender.ARGBOrder]
	PixFmtCompositeBGRA32     = PixFmtCompositeRGBA[color.Linear, blender.BGRAOrder]
	PixFmtCompositeABGR32     = PixFmtCompositeRGBA[color.Linear, blender.ABGROrder]
	PixFmtCompositeRGBA32SRGB = PixFmtCompositeRGBA[color.SRGB, blender.RGBAOrder]
)

// NewPixFmtCompositeRGBA32 creates a new composite RGBA32 pixel format
func NewPixFmtCompositeRGBA32(rbuf *buffer.RenderingBufferU8, op blender.CompOp) *PixFmtCompositeRGBA32 {
	return NewPixFmtCompositeRGBA[color.Linear, blender.RGBAOrder](rbuf, op)
}

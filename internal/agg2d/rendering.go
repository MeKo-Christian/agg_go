// Package agg2d rendering pipeline for AGG2D high-level interface.
// This file contains rendering pipeline methods and functionality.
package agg2d

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/rasterizer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

// Rendering methods

// renderFill renders the current path as a filled shape
func (agg2d *Agg2D) renderFill() {
	if agg2d.rasterizer == nil || agg2d.path == nil || agg2d.scanline == nil {
		return
	}

	// Reset rasterizer for new path
	agg2d.rasterizer.Reset()

	// Apply fill rule (even-odd or non-zero winding)
	if agg2d.evenOddFlag {
		agg2d.rasterizer.FillingRule(basics.FillEvenOdd)
	} else {
		agg2d.rasterizer.FillingRule(basics.FillNonZero)
	}

	// Create transformed curve converter
	transformedPath := conv.NewConvTransform(agg2d.convCurve, agg2d.transform)

	// Add path vertices to rasterizer
	transformedPath.Rewind(0)
	for {
		x, y, cmd := transformedPath.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		agg2d.rasterizer.AddVertex(x, y, uint32(cmd))
	}

	// Render with appropriate color/gradient
	if agg2d.fillGradientFlag == Solid {
		agg2d.renderSolidFill()
	} else {
		agg2d.renderGradientFill()
	}
}

// renderStroke renders the current path as a stroked outline
func (agg2d *Agg2D) renderStroke() {
	if agg2d.rasterizer == nil || agg2d.path == nil || agg2d.convStroke == nil || agg2d.scanline == nil {
		return
	}

	// Reset rasterizer for new path
	agg2d.rasterizer.Reset()

	// Always use non-zero fill rule for strokes
	agg2d.rasterizer.FillingRule(basics.FillNonZero)

	// Create stroke path (potentially with dashes)
	var strokeSource conv.VertexSource
	if agg2d.convDash != nil {
		// Use dashed stroke
		strokeSource = conv.NewConvTransform(agg2d.convDash, agg2d.transform)
	} else {
		// Use regular stroke
		strokeSource = conv.NewConvTransform(agg2d.convStroke, agg2d.transform)
	}

	// Add stroked path vertices to rasterizer
	strokeSource.Rewind(0)
	for {
		x, y, cmd := strokeSource.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		agg2d.rasterizer.AddVertex(x, y, uint32(cmd))
	}

	// Render with appropriate color/gradient
	if agg2d.lineGradientFlag == Solid {
		agg2d.renderSolidStroke()
	} else {
		agg2d.renderGradientStroke()
	}
}

// renderFillWithLineColor renders the current path filled with line color
func (agg2d *Agg2D) renderFillWithLineColor() {
	if agg2d.rasterizer == nil || agg2d.path == nil || agg2d.scanline == nil {
		return
	}

	// Reset rasterizer for new path
	agg2d.rasterizer.Reset()

	// Apply fill rule (even-odd or non-zero winding)
	if agg2d.evenOddFlag {
		agg2d.rasterizer.FillingRule(basics.FillEvenOdd)
	} else {
		agg2d.rasterizer.FillingRule(basics.FillNonZero)
	}

	// Create transformed curve converter
	transformedPath := conv.NewConvTransform(agg2d.convCurve, agg2d.transform)

	// Add path vertices to rasterizer
	transformedPath.Rewind(0)
	for {
		x, y, cmd := transformedPath.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		agg2d.rasterizer.AddVertex(x, y, uint32(cmd))
	}

	// Render using line color instead of fill color
	if agg2d.lineGradientFlag == Solid {
		agg2d.renderSolidFillWithColor(agg2d.lineColor)
	} else {
		agg2d.renderGradientFillWithLineGradient()
	}
}

// renderSolidFill renders solid fill using current fill color
func (agg2d *Agg2D) renderSolidFill() {
	agg2d.renderSolidFillWithColor(agg2d.fillColor)
}

// renderSolidFillWithColor renders solid fill using specified color
func (agg2d *Agg2D) renderSolidFillWithColor(c Color) {
	if agg2d.renBase == nil {
		return
	}

	// Convert Color to internal color format
	internalColor := color.RGBA8[color.Linear]{R: c[0], G: c[1], B: c[2], A: c[3]}

	// Create solid renderer
	renSolid := renscan.NewRendererScanlineAASolidWithColor(agg2d.renBase, internalColor)

	// Render scanlines
	scanlineRender(agg2d.rasterizer, agg2d.scanline, renSolid)
}

// renderSolidStroke renders solid stroke using current line color
func (agg2d *Agg2D) renderSolidStroke() {
	if agg2d.renBase == nil {
		return
	}

	// Convert Color to internal color format
	internalColor := color.RGBA8[color.Linear]{R: agg2d.lineColor[0], G: agg2d.lineColor[1], B: agg2d.lineColor[2], A: agg2d.lineColor[3]}

	// Create solid renderer
	renSolid := renscan.NewRendererScanlineAASolidWithColor(agg2d.renBase, internalColor)

	// Render scanlines
	scanlineRender(agg2d.rasterizer, agg2d.scanline, renSolid)
}

// renderGradientFill renders gradient fill using the appropriate gradient type
func (agg2d *Agg2D) renderGradientFill() {
	if agg2d.fillGradientFlag == Linear {
		agg2d.renderLinearGradientFill(true) // true = use fill gradient settings
	} else if agg2d.fillGradientFlag == Radial {
		agg2d.renderRadialGradientFill(true) // true = use fill gradient settings
	} else {
		// Solid fill fallback
		agg2d.renderSolidFill()
	}
}

// renderLinearGradientFill renders linear gradient fill
func (agg2d *Agg2D) renderLinearGradientFill(useFillGradient bool) {
	if agg2d.renBase == nil || agg2d.spanAllocator == nil {
		return
	}

	// Choose the appropriate gradient settings
	var gradientMatrix *transform.TransAffine
	var gradientColors [256]Color
	var d1, d2 float64

	if useFillGradient {
		gradientMatrix = agg2d.fillGradientMatrix
		gradientColors = agg2d.fillGradient
		d1 = agg2d.fillGradientD1
		d2 = agg2d.fillGradientD2
	} else {
		gradientMatrix = agg2d.lineGradientMatrix
		gradientColors = agg2d.lineGradient
		d1 = agg2d.lineGradientD1
		d2 = agg2d.lineGradientD2
	}

	// Create span interpolator with the gradient transformation matrix
	spanInterpolator := span.NewSpanInterpolatorLinearDefault(gradientMatrix)

	// Convert first and last gradient colors to RGBA8 for span generator
	startColor := color.RGBA8[color.Linear]{R: gradientColors[0][0], G: gradientColors[0][1], B: gradientColors[0][2], A: gradientColors[0][3]}
	endColor := color.RGBA8[color.Linear]{R: gradientColors[255][0], G: gradientColors[255][1], B: gradientColors[255][2], A: gradientColors[255][3]}

	// Create linear gradient span generator
	spanGenerator := span.NewLinearGradientRGBA8(
		spanInterpolator,
		startColor, endColor,
		d1, d2,
		256, // gradient size
	)

	// Render scanlines using the span generator directly
	// Create adapters to bridge interface differences
	rasAdapter := rasterizerAdapter{ras: agg2d.rasterizer}
	slAdapter := &scanlineWrapper{sl: agg2d.scanline}
	renscan.RenderScanlinesAA(rasAdapter, slAdapter, agg2d.renBase, agg2d.spanAllocator, spanGenerator)
}

// renderRadialGradientFill renders radial gradient fill
func (agg2d *Agg2D) renderRadialGradientFill(useFillGradient bool) {
	if agg2d.renBase == nil || agg2d.spanAllocator == nil {
		return
	}

	// Choose the appropriate gradient settings
	var gradientMatrix *transform.TransAffine
	var gradientColors [256]Color
	var d1, d2 float64

	if useFillGradient {
		gradientMatrix = agg2d.fillGradientMatrix
		gradientColors = agg2d.fillGradient
		d1 = agg2d.fillGradientD1
		d2 = agg2d.fillGradientD2
	} else {
		gradientMatrix = agg2d.lineGradientMatrix
		gradientColors = agg2d.lineGradient
		d1 = agg2d.lineGradientD1
		d2 = agg2d.lineGradientD2
	}

	// Create span interpolator with the gradient transformation matrix
	spanInterpolator := span.NewSpanInterpolatorLinearDefault(gradientMatrix)

	// Convert first and last gradient colors to RGBA8 for span generator
	startColor := color.RGBA8[color.Linear]{R: gradientColors[0][0], G: gradientColors[0][1], B: gradientColors[0][2], A: gradientColors[0][3]}
	endColor := color.RGBA8[color.Linear]{R: gradientColors[255][0], G: gradientColors[255][1], B: gradientColors[255][2], A: gradientColors[255][3]}

	// Create radial gradient span generator
	spanGenerator := span.NewRadialGradientRGBA8(
		spanInterpolator,
		startColor, endColor,
		d1, d2,
		256, // gradient size
	)

	// Render scanlines using the span generator directly
	// Create adapters to bridge interface differences
	rasAdapter := rasterizerAdapter{ras: agg2d.rasterizer}
	slAdapter := &scanlineWrapper{sl: agg2d.scanline}
	renscan.RenderScanlinesAA(rasAdapter, slAdapter, agg2d.renBase, agg2d.spanAllocator, spanGenerator)
}

// renderGradientStroke renders gradient stroke using line gradient settings
func (agg2d *Agg2D) renderGradientStroke() {
	if agg2d.lineGradientFlag == Linear {
		agg2d.renderLinearGradientFill(false) // false = use line gradient settings
	} else if agg2d.lineGradientFlag == Radial {
		agg2d.renderRadialGradientFill(false) // false = use line gradient settings
	} else {
		// Solid stroke fallback
		agg2d.renderSolidStroke()
	}
}

// renderGradientFillWithLineGradient renders fill using line gradient settings
func (agg2d *Agg2D) renderGradientFillWithLineGradient() {
	if agg2d.lineGradientFlag == Linear {
		agg2d.renderLinearGradientFill(false) // false = use line gradient settings
	} else if agg2d.lineGradientFlag == Radial {
		agg2d.renderRadialGradientFill(false) // false = use line gradient settings
	} else {
		// Solid fill fallback using line color
		agg2d.renderSolidFillWithColor(agg2d.lineColor)
	}
}

// scanlineRender is a helper function to render scanlines using a renderer
func scanlineRender(rasterizer *rasterizer.RasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl], sl *scanline.ScanlineU8, renderer renscan.RendererInterface[color.RGBA8[color.Linear]]) {
	// Create adapters to bridge interface differences
	rasAdapter := rasterizerAdapter{ras: rasterizer}
	slAdapter := &scanlineWrapper{sl: sl}

	if !rasAdapter.RewindScanlines() {
		return
	}

	// Reset scanline for the rasterizer bounds
	slAdapter.Reset(rasAdapter.MinX(), rasAdapter.MaxX())

	// Prepare the renderer
	renderer.Prepare()

	// Sweep through all scanlines
	for rasAdapter.SweepScanline(slAdapter) {
		renderer.Render(slAdapter)
	}
}

// updateApproximationScales updates the approximation scale for curve converters
// based on the current transformation matrix scaling
func (agg2d *Agg2D) updateApproximationScales() {
	if agg2d.convCurve != nil {
		// Use the world-to-screen scaling factor with the global approximation scale
		scale := agg2d.WorldToScreenScalar(1.0) * ApproxScale

		// Update curve approximation scale
		agg2d.convCurve.SetApproximationScale(scale)
	}

	if agg2d.convStroke != nil {
		// Also update the stroke converter with the same scale for consistency
		scale := agg2d.WorldToScreenScalar(1.0) * ApproxScale
		agg2d.convStroke.SetApproximationScale(scale)
	}
}

// render is the main rendering method that handles both fill and stroke colors
func (agg2d *Agg2D) render(fillColor bool) {
	if fillColor {
		agg2d.renderFill()
	} else {
		agg2d.renderStroke()
	}
}

// updateRasterizerGamma updates the rasterizer gamma correction
func (agg2d *Agg2D) updateRasterizerGamma() {
	if agg2d.rasterizer != nil && agg2d.antiAliasGamma != 1.0 {
		// TODO: Implement gamma correction in rasterizer
		// agg2d.rasterizer.SetGamma(agg2d.antiAliasGamma)
	}
}

// LineWidth sets the line width.
func (agg2d *Agg2D) LineWidth(w float64) {
	agg2d.lineWidth = w
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetWidth(w)
	}
}

// LineCap sets the line cap style.
func (agg2d *Agg2D) LineCap(cap LineCap) {
	agg2d.lineCap = cap
	if agg2d.convStroke != nil {
		switch cap {
		case 0: // CapButt
			agg2d.convStroke.SetLineCap(0) // basics.ButtCap
		case 2: // CapSquare
			agg2d.convStroke.SetLineCap(2) // basics.SquareCap
		case 1: // CapRound
			agg2d.convStroke.SetLineCap(1) // basics.RoundCap
		}
	}
}

// LineJoin sets the line join style.
func (agg2d *Agg2D) LineJoin(join LineJoin) {
	agg2d.lineJoin = join
	if agg2d.convStroke != nil {
		switch join {
		case 0: // JoinMiter
			agg2d.convStroke.SetLineJoin(0) // basics.MiterJoin
		case 1: // JoinRound
			agg2d.convStroke.SetLineJoin(1) // basics.RoundJoin
		case 2: // JoinBevel
			agg2d.convStroke.SetLineJoin(2) // basics.BevelJoin
		}
	}
}

// ResetTransformations resets the transformation matrix to identity.
func (agg2d *Agg2D) ResetTransformations() {
	if agg2d.transform != nil {
		agg2d.transform.Reset()
	}
}

// ImageFilter sets the image filtering method.
func (agg2d *Agg2D) ImageFilter(f ImageFilter) {
	agg2d.imageFilter = f
}

// ImageResample sets the image resampling method.
func (agg2d *Agg2D) ImageResample(r ImageResample) {
	agg2d.imageResample = r
}

// TextAlignment sets text alignment.
func (agg2d *Agg2D) TextAlignment(alignX, alignY TextAlignment) {
	agg2d.textAlignX = alignX
	agg2d.textAlignY = alignY
}

// GetLineWidth returns the current line width
func (agg2d *Agg2D) GetLineWidth() float64 {
	return agg2d.lineWidth
}

// GetLineCap returns the current line cap style
func (agg2d *Agg2D) GetLineCap() LineCap {
	return agg2d.lineCap
}

// GetLineJoin returns the current line join style
func (agg2d *Agg2D) GetLineJoin() LineJoin {
	return agg2d.lineJoin
}

// GetImageFilter returns the current image filter
func (agg2d *Agg2D) GetImageFilter() ImageFilter {
	return agg2d.imageFilter
}

// GetImageResample returns the current image resampling method
func (agg2d *Agg2D) GetImageResample() ImageResample {
	return agg2d.imageResample
}

// GetMasterAlpha returns the current master alpha value
func (agg2d *Agg2D) GetMasterAlpha() float64 {
	return agg2d.masterAlpha
}

// SetMasterAlpha sets the master alpha value
func (agg2d *Agg2D) SetMasterAlpha(alpha float64) {
	if alpha < 0.0 {
		alpha = 0.0
	} else if alpha > 1.0 {
		alpha = 1.0
	}
	agg2d.masterAlpha = alpha
	// TODO: Apply master alpha to renderers
}

// GetAntiAliasGamma returns the current anti-alias gamma value
func (agg2d *Agg2D) GetAntiAliasGamma() float64 {
	return agg2d.antiAliasGamma
}

// SetAntiAliasGamma sets the anti-alias gamma value
func (agg2d *Agg2D) SetAntiAliasGamma(gamma float64) {
	if gamma < 0.1 {
		gamma = 0.1
	} else if gamma > 3.0 {
		gamma = 3.0
	}
	agg2d.antiAliasGamma = gamma
	agg2d.updateRasterizerGamma()
}

// Math helper functions (simplified)
func cos(x float64) float64 {
	return math.Cos(x)
}

func sin(x float64) float64 {
	return math.Sin(x)
}

package rasterizer

import (
	"testing"

	"agg_go/internal/basics"
)

func TestNewRasterizerScanlineAANoGamma(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)

	if r == nil {
		t.Fatal("Expected non-nil rasterizer")
	}

	if r.fillingRule != basics.FillNonZero {
		t.Error("Expected default filling rule to be FillNonZero")
	}

	if !r.autoClose {
		t.Error("Expected autoClose to be true by default")
	}

	if r.status != StatusInitial {
		t.Error("Expected initial status to be StatusInitial")
	}
}

func TestRasterizerScanlineAANoGammaApplyGamma(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)

	// Test that ApplyGamma returns the input unchanged (no gamma correction)
	testValues := []uint32{0, 50, 100, 128, 200, 255}
	for _, val := range testValues {
		result := r.ApplyGamma(val)
		if result != val {
			t.Errorf("Expected ApplyGamma(%d) = %d, got %d", val, val, result)
		}
	}
}

func TestRasterizerScanlineAANoGammaReset(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)

	// Change state
	r.status = StatusMoveTo
	r.startX = 100
	r.startY = 200

	// Reset should restore initial state
	r.Reset()

	if r.status != StatusInitial {
		t.Error("Expected status to be StatusInitial after reset")
	}
}

func TestRasterizerScanlineAANoGammaFillingRule(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)

	// Test setting different filling rules
	r.FillingRule(basics.FillEvenOdd)
	if r.fillingRule != basics.FillEvenOdd {
		t.Error("Expected filling rule to be FillEvenOdd")
	}

	r.FillingRule(basics.FillNonZero)
	if r.fillingRule != basics.FillNonZero {
		t.Error("Expected filling rule to be FillNonZero")
	}
}

func TestRasterizerScanlineAANoGammaAutoClose(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)

	// Test auto close setting
	r.AutoClose(false)
	if r.autoClose {
		t.Error("Expected autoClose to be false")
	}

	r.AutoClose(true)
	if !r.autoClose {
		t.Error("Expected autoClose to be true")
	}
}

func TestRasterizerScanlineAANoGammaMoveTo(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.clipper = clipper

	// Test integer MoveTo
	x, y := 100, 200
	r.MoveTo(x, y)

	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo after MoveTo")
	}

	if r.startX != x || r.startY != y {
		t.Errorf("Expected start position (%d, %d), got (%d, %d)", x, y, r.startX, r.startY)
	}
}

func TestRasterizerScanlineAANoGammaMoveToD(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.clipper = clipper

	// Test floating point MoveTo
	x, y := 10.5, 20.7
	r.MoveToD(x, y)

	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo after MoveToD")
	}

	expectedX := int(x * basics.PolySubpixelScale)
	expectedY := int(y * basics.PolySubpixelScale)

	if r.startX != expectedX || r.startY != expectedY {
		t.Errorf("Expected start position (%d, %d), got (%d, %d)", expectedX, expectedY, r.startX, r.startY)
	}
}

func TestRasterizerScanlineAANoGammaLineTo(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.clipper = clipper

	// Must start with MoveTo
	r.MoveTo(0, 0)

	// Test LineTo
	r.LineTo(100, 200)

	if r.status != StatusLineTo {
		t.Error("Expected status to be StatusLineTo after LineTo")
	}
}

func TestRasterizerScanlineAANoGammaClosePolygon(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.clipper = clipper

	// Create a simple path
	r.MoveTo(0, 0)
	r.LineTo(100, 0)
	r.LineTo(100, 100)

	// Close polygon
	r.ClosePolygon()

	if r.status != StatusClosed {
		t.Error("Expected status to be StatusClosed after ClosePolygon")
	}
}

func TestRasterizerScanlineAANoGammaAddVertex(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.clipper = clipper

	// Test MoveTo command
	r.AddVertex(10.0, 20.0, uint32(basics.PathCmdMoveTo))
	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo after MoveTo command")
	}

	// Test LineTo command
	r.AddVertex(30.0, 40.0, uint32(basics.PathCmdLineTo))
	if r.status != StatusLineTo {
		t.Error("Expected status to be StatusLineTo after LineTo command")
	}

	// Test Close command
	r.AddVertex(0.0, 0.0, uint32(basics.PathFlagsClose))
	if r.status != StatusClosed {
		t.Error("Expected status to be StatusClosed after Close command")
	}
}

func TestRasterizerScanlineAANoGammaEdge(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.clipper = clipper

	// Test Edge with integer coordinates
	r.Edge(0, 0, 100, 100)
	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo after Edge")
	}

	// Test EdgeD with floating point coordinates
	r.EdgeD(10.5, 20.5, 30.7, 40.9)
	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo after EdgeD")
	}
}

func TestRasterizerScanlineAANoGammaCalculateAlpha(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)

	tests := []struct {
		area     int
		expected uint32
	}{
		{0, 0},
		{AAScale << (basics.PolySubpixelShift*2 + 1 - AAShift), AAScale - 1}, // Max value
		{-100, 0}, // Negative area
	}

	for _, test := range tests {
		result := r.CalculateAlpha(test.area)
		if result > AAMask {
			t.Errorf("CalculateAlpha(%d) returned %d, which exceeds AAMask (%d)", test.area, result, AAMask)
		}
	}
}

func TestRasterizerScanlineAANoGammaEvenOddFilling(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.FillingRule(basics.FillEvenOdd)

	// Test even-odd specific calculation
	area := AAScale2 << (basics.PolySubpixelShift*2 + 1 - AAShift)
	alpha := r.CalculateAlpha(area)

	// For even-odd, should handle wrap-around
	if alpha > AAMask {
		t.Errorf("Even-odd alpha calculation failed: got %d, max should be %d", alpha, AAMask)
	}
}

func TestRasterizerScanlineAANoGammaRewindScanlines(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.clipper = clipper

	// Empty rasterizer should return false
	if r.RewindScanlines() {
		t.Error("Expected RewindScanlines to return false for empty rasterizer")
	}

	// Add some geometry
	r.MoveTo(0, 0)
	r.LineTo(100*basics.PolySubpixelScale, 0)
	r.LineTo(100*basics.PolySubpixelScale, 100*basics.PolySubpixelScale)
	r.ClosePolygon()

	// Now should return true (though we can't easily test the actual rasterization without more setup)
	// The test here is mainly to ensure the method doesn't panic and handles the basic case
}

func TestRasterizerScanlineAANoGammaNavigateScanline(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.clipper = clipper

	// Empty rasterizer should return false
	if r.NavigateScanline(100) {
		t.Error("Expected NavigateScanline to return false for empty rasterizer")
	}
}

func TestRasterizerScanlineAANoGammaHitTest(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.clipper = clipper

	// Empty rasterizer should return false for any hit test
	if r.HitTest(50, 50) {
		t.Error("Expected HitTest to return false for empty rasterizer")
	}
}

// MockVertexSource for testing AddPath
type MockVertexSource struct {
	vertices []vertex
	index    int
}

type vertex struct {
	x, y float64
	cmd  uint32
}

func (mvs *MockVertexSource) Rewind(pathID uint32) {
	mvs.index = 0
}

func (mvs *MockVertexSource) Vertex(x, y *float64) uint32 {
	if mvs.index >= len(mvs.vertices) {
		return uint32(basics.PathCmdStop)
	}

	v := mvs.vertices[mvs.index]
	*x = v.x
	*y = v.y
	mvs.index++
	return v.cmd
}

func TestRasterizerScanlineAANoGammaAddPath(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
	r.clipper = clipper

	// Create a simple path: MoveTo, LineTo, LineTo, EndPoly+Close
	vs := &MockVertexSource{
		vertices: []vertex{
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{100, 0, uint32(basics.PathCmdLineTo)},
			{100, 100, uint32(basics.PathCmdLineTo)},
			{0, 0, uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)},
		},
	}

	r.AddPath(vs, 0)

	// The path should have been processed successfully
	// The final status depends on the last command, but we mainly test that it doesn't panic
	// and processes all commands without error
}

// Benchmark comparison with gamma version to ensure performance benefit
func BenchmarkRasterizerScanlineAANoGamma_Triangle(b *testing.B) {
	clipper := &MockClip{}

	for i := 0; i < b.N; i++ {
		r := NewRasterizerScanlineAANoGamma[*MockClip](1024)
		r.clipper = clipper

		// Draw a triangle
		r.MoveTo(0, 0)
		r.LineTo(100*basics.PolySubpixelScale, 0)
		r.LineTo(50*basics.PolySubpixelScale, 100*basics.PolySubpixelScale)
		r.ClosePolygon()

		// Test alpha calculation (main performance difference)
		r.CalculateAlpha(1000)
	}
}

func BenchmarkRasterizerScanlineAANoGamma_ApplyGamma(b *testing.B) {
	r := NewRasterizerScanlineAANoGamma[*MockClip](1024)

	for i := 0; i < b.N; i++ {
		r.ApplyGamma(uint32(i % 256))
	}
}

// Test HitTestScanline implementation
func TestHitTestScanline(t *testing.T) {
	ht := &HitTestScanline{targetX: 50, hit: false}

	// Test ResetSpans
	ht.hit = true
	ht.ResetSpans()
	if ht.hit {
		t.Error("Expected hit to be false after ResetSpans")
	}

	// Test AddCell - hit case
	ht.AddCell(50, 100)
	if !ht.hit {
		t.Error("Expected hit to be true after AddCell with matching X")
	}

	// Test AddCell - miss case
	ht.ResetSpans()
	ht.AddCell(60, 100)
	if ht.hit {
		t.Error("Expected hit to be false after AddCell with non-matching X")
	}

	// Test AddSpan - hit case
	ht.ResetSpans()
	ht.AddSpan(40, 20, 100) // span from 40 to 60, includes 50
	if !ht.hit {
		t.Error("Expected hit to be true after AddSpan containing target")
	}

	// Test AddSpan - miss case
	ht.ResetSpans()
	ht.AddSpan(60, 10, 100) // span from 60 to 70, doesn't include 50
	if ht.hit {
		t.Error("Expected hit to be false after AddSpan not containing target")
	}

	// Test NumSpans
	ht.hit = true
	if ht.NumSpans() != 1 {
		t.Error("Expected NumSpans to return 1 when hit is true")
	}

	ht.hit = false
	if ht.NumSpans() != 0 {
		t.Error("Expected NumSpans to return 0 when hit is false")
	}
}

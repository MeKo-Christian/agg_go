package transform

import (
	"testing"
)

func BenchmarkViewportTransform(b *testing.B) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
	v.DeviceViewport(0.0, 0.0, 1024.0, 768.0)

	x, y := 500.0, 500.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tempX, tempY := x, y
		v.Transform(&tempX, &tempY)
	}
}

func BenchmarkViewportInverseTransform(b *testing.B) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
	v.DeviceViewport(0.0, 0.0, 1024.0, 768.0)

	x, y := 512.0, 384.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tempX, tempY := x, y
		v.InverseTransform(&tempX, &tempY)
	}
}

func BenchmarkViewportTransformBatch(b *testing.B) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
	v.DeviceViewport(0.0, 0.0, 1024.0, 768.0)

	// Create a batch of 1000 coordinate pairs
	coords := make([]float64, 2000)
	for i := 0; i < len(coords); i += 2 {
		coords[i] = float64(i/2) * 0.5
		coords[i+1] = float64(i/2) * 0.5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Make a copy to avoid modifying original
		testCoords := make([]float64, len(coords))
		copy(testCoords, coords)
		v.TransformBatch(testCoords)
	}
}

func BenchmarkViewportTransformPoints(b *testing.B) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
	v.DeviceViewport(0.0, 0.0, 1024.0, 768.0)

	// Create a batch of 1000 points
	points := make([]Point, 1000)
	for i := range points {
		points[i].X = float64(i) * 0.5
		points[i].Y = float64(i) * 0.5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Make a copy to avoid modifying original
		testPoints := make([]Point, len(points))
		copy(testPoints, points)
		v.TransformPoints(testPoints)
	}
}

func BenchmarkViewportToAffine(b *testing.B) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
	v.DeviceViewport(0.0, 0.0, 1024.0, 768.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ToAffine()
	}
}

func BenchmarkViewportToAffineCached(b *testing.B) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
	v.DeviceViewport(0.0, 0.0, 1024.0, 768.0)

	// Prime the cache
	v.ToAffine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ToAffine()
	}
}

func BenchmarkViewportZoomIn(b *testing.B) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
	v.DeviceViewport(0.0, 0.0, 1024.0, 768.0)

	centerX, centerY := 512.0, 384.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset viewport for each iteration
		v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
		v.ZoomIn(1.1, centerX, centerY)
	}
}

func BenchmarkViewportPan(b *testing.B) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
	v.DeviceViewport(0.0, 0.0, 1024.0, 768.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset viewport for each iteration
		v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
		v.Pan(10.0, 10.0)
	}
}

func BenchmarkViewportUpdate(b *testing.B) {
	v := NewTransViewport()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.WorldViewport(float64(i), float64(i), float64(i+1000), float64(i+1000))
	}
}

func BenchmarkViewportManager(b *testing.B) {
	vm := NewViewportManager()

	// Create several viewports
	for i := 0; i < 10; i++ {
		name := "viewport" + string(rune('0'+i))
		vm.CreateViewport(name, 0.0, 0.0, 1000.0, 1000.0, 0.0, 0.0, 1024.0, 768.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := "viewport" + string(rune('0'+(i%10)))
		vm.SwitchTo(name)
		vm.GetCurrent()
	}
}

// Benchmark comparison between individual transforms vs batch transforms
func BenchmarkCompareTransformMethods(b *testing.B) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
	v.DeviceViewport(0.0, 0.0, 1024.0, 768.0)

	numPoints := 1000

	b.Run("Individual", func(b *testing.B) {
		coords := make([]float64, numPoints*2)
		for i := 0; i < len(coords); i += 2 {
			coords[i] = float64(i/2) * 0.5
			coords[i+1] = float64(i/2) * 0.5
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testCoords := make([]float64, len(coords))
			copy(testCoords, coords)

			// Transform each point individually
			for j := 0; j < len(testCoords); j += 2 {
				v.Transform(&testCoords[j], &testCoords[j+1])
			}
		}
	})

	b.Run("Batch", func(b *testing.B) {
		coords := make([]float64, numPoints*2)
		for i := 0; i < len(coords); i += 2 {
			coords[i] = float64(i/2) * 0.5
			coords[i+1] = float64(i/2) * 0.5
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testCoords := make([]float64, len(coords))
			copy(testCoords, coords)
			v.TransformBatch(testCoords)
		}
	})
}

// Benchmark different aspect ratio modes
func BenchmarkAspectRatioModes(b *testing.B) {
	b.Run("Stretch", func(b *testing.B) {
		v := NewTransViewport()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v.WorldViewport(0.0, 0.0, 100.0, 100.0)
			v.DeviceViewport(0.0, 0.0, 200.0, 100.0)
			v.PreserveAspectRatio(0.5, 0.5, AspectRatioStretch)
		}
	})

	b.Run("Meet", func(b *testing.B) {
		v := NewTransViewport()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v.WorldViewport(0.0, 0.0, 100.0, 100.0)
			v.DeviceViewport(0.0, 0.0, 200.0, 100.0)
			v.PreserveAspectRatio(0.5, 0.5, AspectRatioMeet)
		}
	})

	b.Run("Slice", func(b *testing.B) {
		v := NewTransViewport()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v.WorldViewport(0.0, 0.0, 100.0, 100.0)
			v.DeviceViewport(0.0, 0.0, 200.0, 100.0)
			v.PreserveAspectRatio(0.5, 0.5, AspectRatioSlice)
		}
	})
}

// Memory allocation benchmarks
func BenchmarkViewportAllocations(b *testing.B) {
	b.Run("NewTransViewport", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := NewTransViewport()
			_ = v
		}
	})

	b.Run("ToAffineAllocation", func(b *testing.B) {
		v := NewTransViewport()
		v.WorldViewport(0.0, 0.0, 1000.0, 1000.0)
		v.DeviceViewport(0.0, 0.0, 1024.0, 768.0)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			affine := v.ToAffine()
			_ = affine
		}
	})
}

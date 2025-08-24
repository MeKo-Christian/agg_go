//go:build freetype

// Package freetype2 provides integration with the enhanced cache manager v2.
// This bridges the new FreeType2 engine with the fman cache system.
package freetype2

import (
	"fmt"

	"agg_go/internal/fonts"
	"agg_go/internal/scanline"
)

// CacheManager2 integrates the FreeType2 engine with the enhanced cache manager.
// This corresponds to AGG's fman::font_cache_manager2 template class.
type CacheManager2 struct {
	fontEngine   FontEngineInterface
	cachedGlyphs *fonts.FmanCachedGlyphs
	currentFont  LoadedFaceInterface // Reference to current font context
	pathAdaptor  interface{}         // Either *path.SerializedIntegerPathAdaptor[int16] or *path.SerializedIntegerPathAdaptor[int32]
	gray8Adaptor *scanline.SerializedScanlinesAdaptorAA[uint8]
	monoAdaptor  *scanline.SerializedScanlinesAdaptorBin
	lastError    error
}

// NewCacheManager2 creates a new cache manager with the specified font engine.
func NewCacheManager2(fontEngine FontEngineInterface) *CacheManager2 {
	cm := &CacheManager2{
		fontEngine:   fontEngine,
		cachedGlyphs: fonts.NewFmanCachedGlyphs(),
	}

	// Initialize adaptors based on engine type
	cm.initializeAdaptors()

	return cm
}

// initializeAdaptors sets up the appropriate adaptors based on the engine type.
func (cm *CacheManager2) initializeAdaptors() {
	if cm.fontEngine.Is32Bit() {
		// 32-bit engine adaptors
		if engine32, ok := cm.fontEngine.(*FontEngineInt32); ok {
			cm.pathAdaptor = engine32.PathAdaptor()
			adaptorTypes := engine32.Gray8Adaptor()
			if adaptorTypes != nil {
				// TODO: Convert to Fman adaptors when available
				// cm.gray8Adaptor = convertToFmanAA(adaptorTypes.Gray8Adaptor)
				// cm.monoAdaptor = convertToFmanBin(adaptorTypes.MonoAdaptor)
			}
		}
	} else {
		// 16-bit engine adaptors
		if engine16, ok := cm.fontEngine.(*FontEngineInt16); ok {
			cm.pathAdaptor = engine16.PathAdaptor()
			adaptorTypes := engine16.Gray8Adaptor()
			if adaptorTypes != nil {
				// TODO: Convert to Fman adaptors when available
				// cm.gray8Adaptor = convertToFmanAA(adaptorTypes.Gray8Adaptor)
				// cm.monoAdaptor = convertToFmanBin(adaptorTypes.MonoAdaptor)
			}
		}
	}
}

// Glyph retrieves a glyph from the cache, loading it if necessary.
// This corresponds to AGG's fman::font_cache_manager2::glyph method.
func (cm *CacheManager2) Glyph(charCode uint32) *fonts.FmanCachedGlyph {
	// Try to find the glyph in the cache first
	if cachedGlyph := cm.cachedGlyphs.FindGlyph(charCode); cachedGlyph != nil {
		return cachedGlyph
	}

	// Glyph not in cache - need to prepare it
	// For this, we need a loaded face - this should be managed by the font selection process
	// This is a simplified version; in practice, font selection would be more complex

	return nil // TODO: Implement glyph loading from font engine
}

// PathAdaptor returns the path adaptor for vector font rendering.
func (cm *CacheManager2) PathAdaptor() interface{} {
	return cm.pathAdaptor
}

// Gray8Adaptor returns the gray8 scanline adaptor.
func (cm *CacheManager2) Gray8Adaptor() *scanline.SerializedScanlinesAdaptorAA[uint8] {
	return cm.gray8Adaptor
}

// MonoAdaptor returns the mono scanline adaptor.
func (cm *CacheManager2) MonoAdaptor() *scanline.SerializedScanlinesAdaptorBin {
	return cm.monoAdaptor
}

// InitEmbeddedAdaptors initializes the embedded adaptors for a specific glyph.
// This corresponds to AGG's init_embedded_adaptors method.
func (cm *CacheManager2) InitEmbeddedAdaptors(glyph *fonts.FmanCachedGlyph, x, y float64) {
	if glyph == nil {
		return
	}

	switch glyph.DataType {
	case fonts.FmanGlyphDataGray8:
		// Initialize gray8 adaptor with glyph data
		// TODO: Implement when FmanSerializedScanlinesAdaptorAA is available
		// cm.gray8Adaptor = fonts.NewFmanSerializedScanlinesAdaptorAA(glyph.Data, glyph.Bounds, x, y)

	case fonts.FmanGlyphDataMono:
		// Initialize mono adaptor with glyph data
		// TODO: Implement when FmanSerializedScanlinesAdaptorBin is available
		// cm.monoAdaptor = fonts.NewFmanSerializedScanlinesAdaptorBin(glyph.Data, glyph.Bounds, x, y)

	case fonts.FmanGlyphDataOutline:
		// For outline glyphs, the path data should be available in the path adaptor
		// This is handled by the font engine's outline decomposition
		break
	}
}

// AddKerning adds kerning adjustment between two characters.
// This uses the font engine's kerning support.
func (cm *CacheManager2) AddKerning(x, y *float64, first, second uint32) {
	// This requires access to a loaded face
	// In practice, this would be called on the currently active font face
	// TODO: Implement proper font face management and kerning lookup
}

// LastError returns the last error that occurred.
func (cm *CacheManager2) LastError() error {
	return cm.lastError
}

// Close cleans up resources used by the cache manager.
func (cm *CacheManager2) Close() error {
	// Clean up adaptors
	cm.pathAdaptor = nil
	cm.gray8Adaptor = nil
	cm.monoAdaptor = nil

	// Clean up cached glyphs
	if cm.cachedGlyphs != nil {
		// TODO: Implement proper cleanup when available
		cm.cachedGlyphs = nil
	}

	// Clean up font engine
	if cm.fontEngine != nil {
		return cm.fontEngine.Close()
	}

	return nil
}

// FontManager provides a high-level interface for font management with FreeType2.
// This combines font engine creation, face loading, and cache management.
type FontManager struct {
	engines       map[string]FontEngineInterface
	cacheManager  *CacheManager2
	currentFont   string
	defaultEngine string
}

// NewFontManager creates a new font manager with both 16-bit and 32-bit engines.
func NewFontManager() (*FontManager, error) {
	// Create both engine types
	engine16, err := NewFontEngineInt16Default()
	if err != nil {
		return nil, err
	}

	engine32, err := NewFontEngineInt32Default()
	if err != nil {
		engine16.Close()
		return nil, err
	}

	fm := &FontManager{
		engines: map[string]FontEngineInterface{
			"int16": engine16,
			"int32": engine32,
		},
		defaultEngine: "int16", // Start with compact 16-bit engine
	}

	// Initialize cache manager with default engine
	fm.cacheManager = NewCacheManager2(fm.engines[fm.defaultEngine])

	return fm, nil
}

// LoadFont loads a font file and returns a loaded face interface.
func (fm *FontManager) LoadFont(fileName string, preferredEngine string) (LoadedFaceInterface, error) {
	// Select the appropriate engine
	engineKey := preferredEngine
	if engineKey == "" {
		engineKey = fm.defaultEngine
	}

	engine, exists := fm.engines[engineKey]
	if !exists {
		engineKey = fm.defaultEngine
		engine = fm.engines[engineKey]
	}

	// Load the font face
	loadedFace, err := engine.LoadFaceFile(fileName)
	if err != nil {
		return nil, err
	}

	fm.currentFont = fileName

	// Update cache manager if we switched engines
	if engineKey != fm.defaultEngine {
		fm.cacheManager = NewCacheManager2(engine)
	}

	return loadedFace, nil
}

// LoadFontFromMemory loads a font from memory buffer.
func (fm *FontManager) LoadFontFromMemory(buffer []byte, preferredEngine string) (LoadedFaceInterface, error) {
	// Select the appropriate engine
	engineKey := preferredEngine
	if engineKey == "" {
		engineKey = fm.defaultEngine
	}

	engine, exists := fm.engines[engineKey]
	if !exists {
		engineKey = fm.defaultEngine
		engine = fm.engines[engineKey]
	}

	// Load the font face
	loadedFace, err := engine.LoadFace(buffer, uint(len(buffer)))
	if err != nil {
		return nil, err
	}

	fm.currentFont = "memory"

	// Update cache manager if we switched engines
	if engineKey != fm.defaultEngine {
		fm.cacheManager = NewCacheManager2(engine)
	}

	return loadedFace, nil
}

// GetCacheManager returns the current cache manager.
func (fm *FontManager) GetCacheManager() *CacheManager2 {
	return fm.cacheManager
}

// SwitchEngine switches to a different engine type and updates the cache manager.
func (fm *FontManager) SwitchEngine(engineType string) error {
	if _, exists := fm.engines[engineType]; !exists {
		return fmt.Errorf("unknown engine type: %s", engineType)
	}

	if engineType != fm.defaultEngine {
		fm.defaultEngine = engineType
		fm.cacheManager = NewCacheManager2(fm.engines[engineType])
	}

	return nil
}

// Close cleans up all resources used by the font manager.
func (fm *FontManager) Close() error {
	// Close cache manager
	if fm.cacheManager != nil {
		fm.cacheManager.Close()
		fm.cacheManager = nil
	}

	// Close all engines
	for _, engine := range fm.engines {
		if engine != nil {
			engine.Close()
		}
	}
	fm.engines = nil

	return nil
}

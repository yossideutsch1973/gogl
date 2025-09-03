package resource

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// TextureFormat represents the internal format of a texture
type TextureFormat uint32

const (
	FormatRGB    TextureFormat = gl.RGB
	FormatRGBA   TextureFormat = gl.RGBA
	FormatRed    TextureFormat = gl.RED
	FormatRG     TextureFormat = gl.RG
	FormatDepth  TextureFormat = gl.DEPTH_COMPONENT
)

// TextureFilter represents texture filtering modes
type TextureFilter int32

const (
	FilterNearest              TextureFilter = gl.NEAREST
	FilterLinear               TextureFilter = gl.LINEAR
	FilterNearestMipmapNearest TextureFilter = gl.NEAREST_MIPMAP_NEAREST
	FilterLinearMipmapNearest  TextureFilter = gl.LINEAR_MIPMAP_NEAREST
	FilterNearestMipmapLinear  TextureFilter = gl.NEAREST_MIPMAP_LINEAR
	FilterLinearMipmapLinear   TextureFilter = gl.LINEAR_MIPMAP_LINEAR
)

// TextureWrap represents texture wrapping modes
type TextureWrap int32

const (
	WrapRepeat         TextureWrap = gl.REPEAT
	WrapMirroredRepeat TextureWrap = gl.MIRRORED_REPEAT
	WrapClampToEdge    TextureWrap = gl.CLAMP_TO_EDGE
	WrapClampToBorder  TextureWrap = gl.CLAMP_TO_BORDER
)

// TextureConfig holds texture configuration parameters
type TextureConfig struct {
	MinFilter     TextureFilter
	MagFilter     TextureFilter
	WrapS         TextureWrap
	WrapT         TextureWrap
	GenerateMipmap bool
}

// DefaultTextureConfig returns default texture configuration
func DefaultTextureConfig() TextureConfig {
	return TextureConfig{
		MinFilter:     FilterLinear,
		MagFilter:     FilterLinear,
		WrapS:         WrapRepeat,
		WrapT:         WrapRepeat,
		GenerateMipmap: false,
	}
}

// Texture2D represents a 2D texture
type Texture2D struct {
	ID     uint32
	Width  int32
	Height int32
	Format TextureFormat
	Config TextureConfig
}

// NewTexture2D creates a new 2D texture
func NewTexture2D(width, height int32, format TextureFormat, config TextureConfig) (*Texture2D, error) {
	var id uint32
	gl.GenTextures(1, &id)
	if id == 0 {
		return nil, fmt.Errorf("failed to generate texture")
	}

	texture := &Texture2D{
		ID:     id,
		Width:  width,
		Height: height,
		Format: format,
		Config: config,
	}

	// Configure texture
	texture.Bind(0)
	texture.applyConfig()
	texture.Unbind()

	return texture, nil
}

// NewTexture2DFromData creates a texture from raw data
func NewTexture2DFromData(width, height int32, format TextureFormat, data unsafe.Pointer, config TextureConfig) (*Texture2D, error) {
	texture, err := NewTexture2D(width, height, format, config)
	if err != nil {
		return nil, err
	}

	texture.SetData(data)
	return texture, nil
}

// LoadTexture2D loads a texture from a file
func LoadTexture2D(filepath string, config TextureConfig) (*Texture2D, error) {
	// Open file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open texture file: %w", err)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Convert to RGBA
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	// Create texture
	width := int32(rgba.Bounds().Dx())
	height := int32(rgba.Bounds().Dy())

	texture, err := NewTexture2D(width, height, FormatRGBA, config)
	if err != nil {
		return nil, err
	}

	// Upload data
	texture.SetData(gl.Ptr(rgba.Pix))

	return texture, nil
}

// Bind binds the texture to a texture unit
func (t *Texture2D) Bind(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
}

// Unbind unbinds the texture
func (t *Texture2D) Unbind() {
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// SetData sets the texture data
func (t *Texture2D) SetData(data unsafe.Pointer) {
	t.Bind(0)
	
	// Determine data format based on internal format
	var dataFormat uint32
	switch t.Format {
	case FormatRGB:
		dataFormat = gl.RGB
	case FormatRGBA:
		dataFormat = gl.RGBA
	case FormatRed:
		dataFormat = gl.RED
	case FormatRG:
		dataFormat = gl.RG
	case FormatDepth:
		dataFormat = gl.DEPTH_COMPONENT
	default:
		dataFormat = gl.RGBA
	}

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		int32(t.Format),
		t.Width,
		t.Height,
		0,
		dataFormat,
		gl.UNSIGNED_BYTE,
		data,
	)

	if t.Config.GenerateMipmap {
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}

	t.Unbind()
}

// SetSubData updates a portion of the texture
func (t *Texture2D) SetSubData(x, y, width, height int32, data unsafe.Pointer) {
	t.Bind(0)

	// Determine data format
	var dataFormat uint32
	switch t.Format {
	case FormatRGB:
		dataFormat = gl.RGB
	case FormatRGBA:
		dataFormat = gl.RGBA
	case FormatRed:
		dataFormat = gl.RED
	case FormatRG:
		dataFormat = gl.RG
	default:
		dataFormat = gl.RGBA
	}

	gl.TexSubImage2D(
		gl.TEXTURE_2D,
		0,
		x, y,
		width, height,
		dataFormat,
		gl.UNSIGNED_BYTE,
		data,
	)

	if t.Config.GenerateMipmap {
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}

	t.Unbind()
}

// applyConfig applies texture configuration
func (t *Texture2D) applyConfig() {
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, int32(t.Config.MinFilter))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, int32(t.Config.MagFilter))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, int32(t.Config.WrapS))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, int32(t.Config.WrapT))
}

// SetFilter sets texture filtering
func (t *Texture2D) SetFilter(min, mag TextureFilter) {
	t.Config.MinFilter = min
	t.Config.MagFilter = mag
	t.Bind(0)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, int32(min))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, int32(mag))
	t.Unbind()
}

// SetWrap sets texture wrapping
func (t *Texture2D) SetWrap(s, t_ TextureWrap) {
	t.Config.WrapS = s
	t.Config.WrapT = t_
	t.Bind(0)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, int32(s))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, int32(t_))
	t.Unbind()
}

// GenerateMipmaps generates mipmaps for the texture
func (t *Texture2D) GenerateMipmaps() {
	t.Bind(0)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	t.Config.GenerateMipmap = true
	t.Unbind()
}

// Delete deletes the texture
func (t *Texture2D) Delete() {
	if t.ID != 0 {
		gl.DeleteTextures(1, &t.ID)
		t.ID = 0
	}
}

// TextureArray represents a 2D texture array
type TextureArray struct {
	ID     uint32
	Width  int32
	Height int32
	Layers int32
	Format TextureFormat
	Config TextureConfig
}

// NewTextureArray creates a new texture array
func NewTextureArray(width, height, layers int32, format TextureFormat, config TextureConfig) (*TextureArray, error) {
	var id uint32
	gl.GenTextures(1, &id)
	if id == 0 {
		return nil, fmt.Errorf("failed to generate texture array")
	}

	texture := &TextureArray{
		ID:     id,
		Width:  width,
		Height: height,
		Layers: layers,
		Format: format,
		Config: config,
	}

	// Allocate storage
	texture.Bind(0)
	
	gl.TexImage3D(
		gl.TEXTURE_2D_ARRAY,
		0,
		int32(format),
		width,
		height,
		layers,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		nil,
	)

	// Apply configuration
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MIN_FILTER, int32(config.MinFilter))
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAG_FILTER, int32(config.MagFilter))
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_S, int32(config.WrapS))
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_T, int32(config.WrapT))

	texture.Unbind()

	return texture, nil
}

// Bind binds the texture array
func (ta *TextureArray) Bind(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, ta.ID)
}

// Unbind unbinds the texture array
func (ta *TextureArray) Unbind() {
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, 0)
}

// SetLayerData sets data for a specific layer
func (ta *TextureArray) SetLayerData(layer int32, data unsafe.Pointer) {
	ta.Bind(0)

	var dataFormat uint32
	switch ta.Format {
	case FormatRGB:
		dataFormat = gl.RGB
	case FormatRGBA:
		dataFormat = gl.RGBA
	default:
		dataFormat = gl.RGBA
	}

	gl.TexSubImage3D(
		gl.TEXTURE_2D_ARRAY,
		0,
		0, 0, layer,
		ta.Width, ta.Height, 1,
		dataFormat,
		gl.UNSIGNED_BYTE,
		data,
	)

	if ta.Config.GenerateMipmap {
		gl.GenerateMipmap(gl.TEXTURE_2D_ARRAY)
	}

	ta.Unbind()
}

// Delete deletes the texture array
func (ta *TextureArray) Delete() {
	if ta.ID != 0 {
		gl.DeleteTextures(1, &ta.ID)
		ta.ID = 0
	}
}

// TextureManager manages texture resources
type TextureManager struct {
	textures map[string]*Texture2D
}

// NewTextureManager creates a new texture manager
func NewTextureManager() *TextureManager {
	return &TextureManager{
		textures: make(map[string]*Texture2D),
	}
}

// Load loads a texture from file
func (tm *TextureManager) Load(name, filepath string, config TextureConfig) (*Texture2D, error) {
	// Check if already loaded
	if tex, ok := tm.textures[name]; ok {
		return tex, nil
	}

	// Load texture
	texture, err := LoadTexture2D(filepath, config)
	if err != nil {
		return nil, err
	}

	tm.textures[name] = texture
	return texture, nil
}

// Get retrieves a texture by name
func (tm *TextureManager) Get(name string) *Texture2D {
	return tm.textures[name]
}

// Delete removes a texture
func (tm *TextureManager) Delete(name string) {
	if tex, ok := tm.textures[name]; ok {
		tex.Delete()
		delete(tm.textures, name)
	}
}

// Clear deletes all textures
func (tm *TextureManager) Clear() {
	for _, tex := range tm.textures {
		tex.Delete()
	}
	tm.textures = make(map[string]*Texture2D)
}
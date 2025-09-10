package pipeline

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/yossideutsch/gogl/pkg/shader"
)

// BlendFunc represents OpenGL blend functions
type BlendFunc uint32

const (
	BlendZero             BlendFunc = gl.ZERO
	BlendOne              BlendFunc = gl.ONE
	BlendSrcColor         BlendFunc = gl.SRC_COLOR
	BlendOneMinusSrcColor BlendFunc = gl.ONE_MINUS_SRC_COLOR
	BlendDstColor         BlendFunc = gl.DST_COLOR
	BlendOneMinusDstColor BlendFunc = gl.ONE_MINUS_DST_COLOR
	BlendSrcAlpha         BlendFunc = gl.SRC_ALPHA
	BlendOneMinusSrcAlpha BlendFunc = gl.ONE_MINUS_SRC_ALPHA
	BlendDstAlpha         BlendFunc = gl.DST_ALPHA
	BlendOneMinusDstAlpha BlendFunc = gl.ONE_MINUS_DST_ALPHA
)

// CullFace represents face culling modes
type CullFace uint32

const (
	CullNone  CullFace = 0
	CullFront CullFace = gl.FRONT
	CullBack  CullFace = gl.BACK
)

// DepthFunc represents depth comparison functions
type DepthFunc uint32

const (
	DepthNever    DepthFunc = gl.NEVER
	DepthLess     DepthFunc = gl.LESS
	DepthEqual    DepthFunc = gl.EQUAL
	DepthLessEq   DepthFunc = gl.LEQUAL
	DepthGreater  DepthFunc = gl.GREATER
	DepthNotEqual DepthFunc = gl.NOTEQUAL
	DepthGreaterEq DepthFunc = gl.GEQUAL
	DepthAlways   DepthFunc = gl.ALWAYS
)

// Primitive represents OpenGL primitive types
type Primitive uint32

const (
	Points        Primitive = gl.POINTS
	Lines         Primitive = gl.LINES
	LineLoop      Primitive = gl.LINE_LOOP
	LineStrip     Primitive = gl.LINE_STRIP
	Triangles     Primitive = gl.TRIANGLES
	TriangleStrip Primitive = gl.TRIANGLE_STRIP
	TriangleFan   Primitive = gl.TRIANGLE_FAN
)

// State represents the complete OpenGL rendering state
type State struct {
	// Shader program
	Program *shader.Program

	// Blending
	BlendEnabled bool
	BlendSrc     BlendFunc
	BlendDst     BlendFunc

	// Depth testing
	DepthEnabled bool
	DepthWrite   bool
	DepthFunc    DepthFunc

	// Face culling
	CullEnabled bool
	CullFace    CullFace

	// Viewport
	ViewportX      int32
	ViewportY      int32
	ViewportWidth  int32
	ViewportHeight int32

	// Polygon mode
	WireframeMode bool

	// Primitive type
	Primitive Primitive
}

// DefaultState returns a sensible default pipeline state
func DefaultState() *State {
	return &State{
		BlendEnabled: false,
		BlendSrc:     BlendSrcAlpha,
		BlendDst:     BlendOneMinusSrcAlpha,

		DepthEnabled: true,
		DepthWrite:   true,
		DepthFunc:    DepthLess,

		CullEnabled: true,
		CullFace:    CullBack,

		ViewportX:      0,
		ViewportY:      0,
		ViewportWidth:  800,
		ViewportHeight: 600,

		WireframeMode: false,
		Primitive:     Triangles,
	}
}

// Pipeline manages the OpenGL rendering pipeline state
type Pipeline struct {
	currentState *State
	stateStack   []*State
	// Cache to avoid redundant state changes
	lastProgramID  uint32
	lastBlendState bool
	lastDepthState bool
	lastCullState  bool
}

// New creates a new rendering pipeline
func New() *Pipeline {
	return &Pipeline{
		currentState: DefaultState(),
		stateStack:   make([]*State, 0),
	}
}

// SetState sets the complete pipeline state with optimized state changes
func (p *Pipeline) SetState(state *State) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	// Apply shader program only if changed
	if state.Program != nil && (p.lastProgramID != state.Program.ID) {
		state.Program.Use()
		p.lastProgramID = state.Program.ID
	}

	// Apply blending state only if changed
	if p.lastBlendState != state.BlendEnabled {
		if state.BlendEnabled {
			gl.Enable(gl.BLEND)
			gl.BlendFunc(uint32(state.BlendSrc), uint32(state.BlendDst))
		} else {
			gl.Disable(gl.BLEND)
		}
		p.lastBlendState = state.BlendEnabled
	} else if state.BlendEnabled {
		// Update blend function even if blend is already enabled
		gl.BlendFunc(uint32(state.BlendSrc), uint32(state.BlendDst))
	}

	// Apply depth state only if changed
	if p.lastDepthState != state.DepthEnabled {
		if state.DepthEnabled {
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(uint32(state.DepthFunc))
			gl.DepthMask(state.DepthWrite)
		} else {
			gl.Disable(gl.DEPTH_TEST)
		}
		p.lastDepthState = state.DepthEnabled
	} else if state.DepthEnabled {
		// Update depth function and mask even if depth test is already enabled
		gl.DepthFunc(uint32(state.DepthFunc))
		gl.DepthMask(state.DepthWrite)
	}

	// Apply culling state only if changed
	cullStateChanged := p.lastCullState != state.CullEnabled
	if cullStateChanged {
		if state.CullEnabled && state.CullFace != CullNone {
			gl.Enable(gl.CULL_FACE)
			gl.CullFace(uint32(state.CullFace))
		} else {
			gl.Disable(gl.CULL_FACE)
		}
		p.lastCullState = state.CullEnabled
	} else if state.CullEnabled && state.CullFace != CullNone {
		// Update cull face even if culling is already enabled
		gl.CullFace(uint32(state.CullFace))
	}

	// Always apply viewport (relatively cheap and may change frequently)
	gl.Viewport(state.ViewportX, state.ViewportY, state.ViewportWidth, state.ViewportHeight)

	// Apply polygon mode
	if state.WireframeMode {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	} else {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}

	p.currentState = state
	return nil
}

// GetState returns the current pipeline state
func (p *Pipeline) GetState() *State {
	return p.currentState
}

// PushState saves the current state on the stack
func (p *Pipeline) PushState() {
	// Create a copy of the current state
	stateCopy := *p.currentState
	p.stateStack = append(p.stateStack, &stateCopy)
}

// PopState restores the previous state from the stack
func (p *Pipeline) PopState() error {
	if len(p.stateStack) == 0 {
		return fmt.Errorf("state stack is empty")
	}

	// Pop the last state
	lastIndex := len(p.stateStack) - 1
	state := p.stateStack[lastIndex]
	p.stateStack = p.stateStack[:lastIndex]

	// Apply the popped state
	return p.SetState(state)
}

// SetProgram sets the shader program with caching
func (p *Pipeline) SetProgram(program *shader.Program) {
	if program != nil && p.lastProgramID != program.ID {
		p.currentState.Program = program
		program.Use()
		p.lastProgramID = program.ID
	} else if program == nil && p.lastProgramID != 0 {
		p.currentState.Program = nil
		p.lastProgramID = 0
	}
}

// SetBlending configures blending
func (p *Pipeline) SetBlending(enabled bool, src, dst BlendFunc) {
	p.currentState.BlendEnabled = enabled
	p.currentState.BlendSrc = src
	p.currentState.BlendDst = dst

	if enabled {
		gl.Enable(gl.BLEND)
		gl.BlendFunc(uint32(src), uint32(dst))
	} else {
		gl.Disable(gl.BLEND)
	}
}

// SetDepthTest configures depth testing
func (p *Pipeline) SetDepthTest(enabled bool, write bool, fn DepthFunc) {
	p.currentState.DepthEnabled = enabled
	p.currentState.DepthWrite = write
	p.currentState.DepthFunc = fn

	if enabled {
		gl.Enable(gl.DEPTH_TEST)
		gl.DepthFunc(uint32(fn))
		gl.DepthMask(write)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
}

// SetCulling configures face culling
func (p *Pipeline) SetCulling(enabled bool, face CullFace) {
	p.currentState.CullEnabled = enabled
	p.currentState.CullFace = face

	if enabled && face != CullNone {
		gl.Enable(gl.CULL_FACE)
		gl.CullFace(uint32(face))
	} else {
		gl.Disable(gl.CULL_FACE)
	}
}

// SetViewport sets the rendering viewport
func (p *Pipeline) SetViewport(x, y, width, height int32) {
	p.currentState.ViewportX = x
	p.currentState.ViewportY = y
	p.currentState.ViewportWidth = width
	p.currentState.ViewportHeight = height
	gl.Viewport(x, y, width, height)
}

// SetWireframe enables or disables wireframe rendering
func (p *Pipeline) SetWireframe(enabled bool) {
	p.currentState.WireframeMode = enabled
	if enabled {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	} else {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}
}

// Clear clears the framebuffer
func (p *Pipeline) Clear(color bool, depth bool, stencil bool) {
	var mask uint32
	if color {
		mask |= gl.COLOR_BUFFER_BIT
	}
	if depth {
		mask |= gl.DEPTH_BUFFER_BIT
	}
	if stencil {
		mask |= gl.STENCIL_BUFFER_BIT
	}
	gl.Clear(mask)
}

// SetClearColor sets the clear color
func (p *Pipeline) SetClearColor(r, g, b, a float32) {
	gl.ClearColor(r, g, b, a)
}

// Builder provides a fluent interface for configuring pipeline state
type Builder struct {
	state *State
}

// NewBuilder creates a new pipeline state builder
func NewBuilder() *Builder {
	return &Builder{
		state: DefaultState(),
	}
}

// WithProgram sets the shader program
func (b *Builder) WithProgram(program *shader.Program) *Builder {
	b.state.Program = program
	return b
}

// WithBlending configures blending
func (b *Builder) WithBlending(enabled bool, src, dst BlendFunc) *Builder {
	b.state.BlendEnabled = enabled
	b.state.BlendSrc = src
	b.state.BlendDst = dst
	return b
}

// WithDepthTest configures depth testing
func (b *Builder) WithDepthTest(enabled bool, write bool, fn DepthFunc) *Builder {
	b.state.DepthEnabled = enabled
	b.state.DepthWrite = write
	b.state.DepthFunc = fn
	return b
}

// WithCulling configures face culling
func (b *Builder) WithCulling(enabled bool, face CullFace) *Builder {
	b.state.CullEnabled = enabled
	b.state.CullFace = face
	return b
}

// WithViewport sets the viewport
func (b *Builder) WithViewport(x, y, width, height int32) *Builder {
	b.state.ViewportX = x
	b.state.ViewportY = y
	b.state.ViewportWidth = width
	b.state.ViewportHeight = height
	return b
}

// WithWireframe enables wireframe mode
func (b *Builder) WithWireframe(enabled bool) *Builder {
	b.state.WireframeMode = enabled
	return b
}

// WithPrimitive sets the primitive type
func (b *Builder) WithPrimitive(primitive Primitive) *Builder {
	b.state.Primitive = primitive
	return b
}

// Build returns the configured state
func (b *Builder) Build() *State {
	return b.state
}

// Validate checks if the current state is valid
func (s *State) Validate() error {
	if s.ViewportWidth <= 0 || s.ViewportHeight <= 0 {
		return fmt.Errorf("invalid viewport dimensions: %dx%d", s.ViewportWidth, s.ViewportHeight)
	}

	if s.BlendEnabled && s.BlendSrc == BlendZero && s.BlendDst == BlendZero {
		return fmt.Errorf("invalid blend function: both source and destination are ZERO")
	}

	return nil
}
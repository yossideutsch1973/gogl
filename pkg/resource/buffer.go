package resource

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// BufferUsage represents how the buffer will be used
type BufferUsage uint32

const (
	StaticDraw  BufferUsage = gl.STATIC_DRAW
	DynamicDraw BufferUsage = gl.DYNAMIC_DRAW
	StreamDraw  BufferUsage = gl.STREAM_DRAW
)

// BufferTarget represents the buffer binding target
type BufferTarget uint32

const (
	ArrayBuffer               BufferTarget = gl.ARRAY_BUFFER
	ElementArrayBuffer        BufferTarget = gl.ELEMENT_ARRAY_BUFFER
	UniformBufferTarget       BufferTarget = gl.UNIFORM_BUFFER
	ShaderStorageBufferTarget BufferTarget = gl.SHADER_STORAGE_BUFFER
)

// Buffer represents an OpenGL buffer object
type Buffer struct {
	ID     uint32
	Target BufferTarget
	Size   int
	Usage  BufferUsage
}

// VertexBuffer represents a buffer for vertex data
type VertexBuffer struct {
	*Buffer
}

// IndexBuffer represents a buffer for index data
type IndexBuffer struct {
	*Buffer
	Count      int
	IndexType  uint32
}

// UniformBuffer represents a buffer for uniform data
type UniformBuffer struct {
	*Buffer
	BindingPoint uint32
}

// ShaderStorageBuffer represents a buffer for shader storage
type ShaderStorageBuffer struct {
	*Buffer
	BindingPoint uint32
}

// createBuffer creates a new OpenGL buffer
func createBuffer(target BufferTarget, data unsafe.Pointer, size int, usage BufferUsage) (*Buffer, error) {
	var id uint32
	gl.GenBuffers(1, &id)
	if id == 0 {
		return nil, fmt.Errorf("failed to generate buffer")
	}

	buffer := &Buffer{
		ID:     id,
		Target: target,
		Size:   size,
		Usage:  usage,
	}

	gl.BindBuffer(uint32(target), id)
	gl.BufferData(uint32(target), size, data, uint32(usage))
	gl.BindBuffer(uint32(target), 0)

	return buffer, nil
}

// NewVertexBuffer creates a new vertex buffer
func NewVertexBuffer(data []float32, usage BufferUsage) (*VertexBuffer, error) {
	size := len(data) * 4 // float32 is 4 bytes
	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = gl.Ptr(data)
	}

	buffer, err := createBuffer(ArrayBuffer, ptr, size, usage)
	if err != nil {
		return nil, err
	}

	return &VertexBuffer{Buffer: buffer}, nil
}

// NewIndexBuffer creates a new index buffer
func NewIndexBuffer(data []uint32, usage BufferUsage) (*IndexBuffer, error) {
	size := len(data) * 4 // uint32 is 4 bytes
	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = gl.Ptr(data)
	}

	buffer, err := createBuffer(ElementArrayBuffer, ptr, size, usage)
	if err != nil {
		return nil, err
	}

	return &IndexBuffer{
		Buffer:    buffer,
		Count:     len(data),
		IndexType: gl.UNSIGNED_INT,
	}, nil
}

// NewIndexBuffer16 creates a new index buffer with 16-bit indices
func NewIndexBuffer16(data []uint16, usage BufferUsage) (*IndexBuffer, error) {
	size := len(data) * 2 // uint16 is 2 bytes
	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = gl.Ptr(data)
	}

	buffer, err := createBuffer(ElementArrayBuffer, ptr, size, usage)
	if err != nil {
		return nil, err
	}

	return &IndexBuffer{
		Buffer:    buffer,
		Count:     len(data),
		IndexType: gl.UNSIGNED_SHORT,
	}, nil
}

// NewUniformBuffer creates a new uniform buffer
func NewUniformBuffer(size int, usage BufferUsage) (*UniformBuffer, error) {
	buffer, err := createBuffer(UniformBufferTarget, nil, size, usage)
	if err != nil {
		return nil, err
	}

	return &UniformBuffer{
		Buffer:       buffer,
		BindingPoint: 0,
	}, nil
}

// NewShaderStorageBuffer creates a new shader storage buffer
func NewShaderStorageBuffer(size int, usage BufferUsage) (*ShaderStorageBuffer, error) {
	buffer, err := createBuffer(ShaderStorageBufferTarget, nil, size, usage)
	if err != nil {
		return nil, err
	}

	return &ShaderStorageBuffer{
		Buffer:       buffer,
		BindingPoint: 0,
	}, nil
}

// Bind binds the buffer
func (b *Buffer) Bind() {
	gl.BindBuffer(uint32(b.Target), b.ID)
}

// Unbind unbinds the buffer
func (b *Buffer) Unbind() {
	gl.BindBuffer(uint32(b.Target), 0)
}

// Update updates the buffer data
func (b *Buffer) Update(offset int, data unsafe.Pointer, size int) error {
	if offset+size > b.Size {
		return fmt.Errorf("data exceeds buffer size")
	}

	b.Bind()
	gl.BufferSubData(uint32(b.Target), offset, size, data)
	b.Unbind()
	return nil
}

// UpdateFloat32 updates the buffer with float32 data
func (v *VertexBuffer) UpdateFloat32(offset int, data []float32) error {
	size := len(data) * 4
	if len(data) == 0 {
		return nil
	}
	return v.Update(offset, gl.Ptr(data), size)
}

// UpdateUint32 updates the index buffer with uint32 data
func (i *IndexBuffer) UpdateUint32(offset int, data []uint32) error {
	size := len(data) * 4
	if len(data) == 0 {
		return nil
	}
	i.Count = len(data)
	return i.Update(offset, gl.Ptr(data), size)
}

// UpdateUint16 updates the index buffer with uint16 data
func (i *IndexBuffer) UpdateUint16(offset int, data []uint16) error {
	size := len(data) * 2
	if len(data) == 0 {
		return nil
	}
	i.Count = len(data)
	i.IndexType = gl.UNSIGNED_SHORT
	return i.Update(offset, gl.Ptr(data), size)
}

// BindBase binds the uniform buffer to a binding point
func (u *UniformBuffer) BindBase(bindingPoint uint32) {
	u.BindingPoint = bindingPoint
	gl.BindBufferBase(gl.UNIFORM_BUFFER, bindingPoint, u.ID)
}

// UpdateData updates uniform buffer data
func (u *UniformBuffer) UpdateData(offset int, data unsafe.Pointer, size int) error {
	return u.Update(offset, data, size)
}

// BindBase binds the shader storage buffer to a binding point
func (s *ShaderStorageBuffer) BindBase(bindingPoint uint32) {
	s.BindingPoint = bindingPoint
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, bindingPoint, s.ID)
}

// UpdateData updates shader storage buffer data
func (s *ShaderStorageBuffer) UpdateData(offset int, data unsafe.Pointer, size int) error {
	return s.Update(offset, data, size)
}

// Delete deletes the buffer
func (b *Buffer) Delete() {
	if b.ID != 0 {
		gl.DeleteBuffers(1, &b.ID)
		b.ID = 0
	}
}

// BufferPool manages a pool of reusable buffers
type BufferPool struct {
	availableBuffers map[BufferTarget][]*Buffer
	inUseBuffers     map[uint32]*Buffer
}

// NewBufferPool creates a new buffer pool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		availableBuffers: make(map[BufferTarget][]*Buffer),
		inUseBuffers:     make(map[uint32]*Buffer),
	}
}

// Acquire gets a buffer from the pool or creates a new one
func (p *BufferPool) Acquire(target BufferTarget, size int, usage BufferUsage) (*Buffer, error) {
	// Check for available buffer of sufficient size
	if buffers, ok := p.availableBuffers[target]; ok {
		for i, buf := range buffers {
			if buf.Size >= size && buf.Usage == usage {
				// Remove from available
				p.availableBuffers[target] = append(buffers[:i], buffers[i+1:]...)
				// Add to in-use
				p.inUseBuffers[buf.ID] = buf
				return buf, nil
			}
		}
	}

	// Create new buffer
	buffer, err := createBuffer(target, nil, size, usage)
	if err != nil {
		return nil, err
	}

	p.inUseBuffers[buffer.ID] = buffer
	return buffer, nil
}

// Release returns a buffer to the pool
func (p *BufferPool) Release(buffer *Buffer) {
	if buffer == nil || buffer.ID == 0 {
		return
	}

	// Remove from in-use
	delete(p.inUseBuffers, buffer.ID)

	// Add to available
	if p.availableBuffers[buffer.Target] == nil {
		p.availableBuffers[buffer.Target] = make([]*Buffer, 0)
	}
	p.availableBuffers[buffer.Target] = append(p.availableBuffers[buffer.Target], buffer)
}

// Clear deletes all buffers in the pool
func (p *BufferPool) Clear() {
	// Delete all available buffers
	for _, buffers := range p.availableBuffers {
		for _, buf := range buffers {
			buf.Delete()
		}
	}
	p.availableBuffers = make(map[BufferTarget][]*Buffer)

	// Delete all in-use buffers
	for _, buf := range p.inUseBuffers {
		buf.Delete()
	}
	p.inUseBuffers = make(map[uint32]*Buffer)
}
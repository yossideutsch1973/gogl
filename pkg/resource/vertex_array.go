package resource

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// AttributeType represents the data type of a vertex attribute
type AttributeType uint32

const (
	Float    AttributeType = gl.FLOAT
	Int      AttributeType = gl.INT
	UInt     AttributeType = gl.UNSIGNED_INT
	Byte     AttributeType = gl.BYTE
	UByte    AttributeType = gl.UNSIGNED_BYTE
	Short    AttributeType = gl.SHORT
	UShort   AttributeType = gl.UNSIGNED_SHORT
)

// VertexAttribute describes a vertex attribute
type VertexAttribute struct {
	Location   uint32
	Size       int32          // Number of components (1, 2, 3, or 4)
	Type       AttributeType
	Normalized bool
	Stride     int32
	Offset     uintptr
	Divisor    uint32        // For instanced rendering
}

// VertexArray represents an OpenGL vertex array object (VAO)
type VertexArray struct {
	ID         uint32
	Attributes []VertexAttribute
	VBO        *VertexBuffer
	IBO        *IndexBuffer
}

// NewVertexArray creates a new vertex array object
func NewVertexArray() (*VertexArray, error) {
	var id uint32
	gl.GenVertexArrays(1, &id)
	if id == 0 {
		return nil, fmt.Errorf("failed to generate vertex array")
	}

	return &VertexArray{
		ID:         id,
		Attributes: make([]VertexAttribute, 0),
	}, nil
}

// Bind binds the vertex array
func (va *VertexArray) Bind() {
	gl.BindVertexArray(va.ID)
}

// Unbind unbinds the vertex array
func (va *VertexArray) Unbind() {
	gl.BindVertexArray(0)
}

// SetVertexBuffer associates a vertex buffer with this VAO
func (va *VertexArray) SetVertexBuffer(vbo *VertexBuffer) {
	va.VBO = vbo
	va.Bind()
	vbo.Bind()
	va.Unbind()
}

// SetIndexBuffer associates an index buffer with this VAO
func (va *VertexArray) SetIndexBuffer(ibo *IndexBuffer) {
	va.IBO = ibo
	va.Bind()
	ibo.Bind()
	va.Unbind()
}

// AddAttribute adds a vertex attribute
func (va *VertexArray) AddAttribute(attr VertexAttribute) {
	va.Attributes = append(va.Attributes, attr)
	
	va.Bind()
	if va.VBO != nil {
		va.VBO.Bind()
	}

	gl.EnableVertexAttribArray(attr.Location)
	
	// Configure the attribute
	switch attr.Type {
	case Float:
		gl.VertexAttribPointer(
			attr.Location,
			attr.Size,
			uint32(attr.Type),
			attr.Normalized,
			attr.Stride,
			gl.PtrOffset(int(attr.Offset)),
		)
	case Int, UInt, Byte, UByte, Short, UShort:
		gl.VertexAttribIPointer(
			attr.Location,
			attr.Size,
			uint32(attr.Type),
			attr.Stride,
			gl.PtrOffset(int(attr.Offset)),
		)
	}

	// Set divisor for instanced rendering
	if attr.Divisor > 0 {
		gl.VertexAttribDivisor(attr.Location, attr.Divisor)
	}

	va.Unbind()
}

// AddFloatAttribute is a convenience method for adding float attributes
func (va *VertexArray) AddFloatAttribute(location uint32, size int32, stride int32, offset uintptr) {
	va.AddAttribute(VertexAttribute{
		Location:   location,
		Size:       size,
		Type:       Float,
		Normalized: false,
		Stride:     stride,
		Offset:     offset,
		Divisor:    0,
	})
}

// Delete deletes the vertex array
func (va *VertexArray) Delete() {
	if va.ID != 0 {
		// Disable all attributes
		va.Bind()
		for _, attr := range va.Attributes {
			gl.DisableVertexAttribArray(attr.Location)
		}
		va.Unbind()

		gl.DeleteVertexArrays(1, &va.ID)
		va.ID = 0
	}
}

// Draw draws the vertex array
func (va *VertexArray) Draw(mode uint32, count int32, offset int32) {
	va.Bind()
	if va.IBO != nil {
		gl.DrawElements(mode, count, va.IBO.IndexType, gl.PtrOffset(int(offset)*4))
	} else {
		gl.DrawArrays(mode, offset, count)
	}
	va.Unbind()
}

// DrawIndexed draws using the index buffer
func (va *VertexArray) DrawIndexed(mode uint32) {
	if va.IBO == nil {
		return
	}
	va.Draw(mode, int32(va.IBO.Count), 0)
}

// DrawInstanced draws multiple instances
func (va *VertexArray) DrawInstanced(mode uint32, count int32, instanceCount int32, offset int32) {
	va.Bind()
	if va.IBO != nil {
		gl.DrawElementsInstanced(mode, count, va.IBO.IndexType, gl.PtrOffset(int(offset)*4), instanceCount)
	} else {
		gl.DrawArraysInstanced(mode, offset, count, instanceCount)
	}
	va.Unbind()
}

// VertexLayout helps build vertex attribute layouts
type VertexLayout struct {
	Attributes []VertexAttribute
	Stride     int32
}

// NewVertexLayout creates a new vertex layout builder
func NewVertexLayout() *VertexLayout {
	return &VertexLayout{
		Attributes: make([]VertexAttribute, 0),
		Stride:     0,
	}
}

// AddFloat adds a float attribute to the layout
func (vl *VertexLayout) AddFloat(location uint32, count int32) *VertexLayout {
	attr := VertexAttribute{
		Location:   location,
		Size:       count,
		Type:       Float,
		Normalized: false,
		Stride:     0, // Will be set when applied
		Offset:     uintptr(vl.Stride),
		Divisor:    0,
	}
	vl.Attributes = append(vl.Attributes, attr)
	vl.Stride += count * 4 // float32 is 4 bytes
	return vl
}

// AddInt adds an integer attribute to the layout
func (vl *VertexLayout) AddInt(location uint32, count int32) *VertexLayout {
	attr := VertexAttribute{
		Location:   location,
		Size:       count,
		Type:       Int,
		Normalized: false,
		Stride:     0, // Will be set when applied
		Offset:     uintptr(vl.Stride),
		Divisor:    0,
	}
	vl.Attributes = append(vl.Attributes, attr)
	vl.Stride += count * 4 // int32 is 4 bytes
	return vl
}

// AddUByte adds an unsigned byte attribute to the layout
func (vl *VertexLayout) AddUByte(location uint32, count int32, normalized bool) *VertexLayout {
	attr := VertexAttribute{
		Location:   location,
		Size:       count,
		Type:       UByte,
		Normalized: normalized,
		Stride:     0, // Will be set when applied
		Offset:     uintptr(vl.Stride),
		Divisor:    0,
	}
	vl.Attributes = append(vl.Attributes, attr)
	vl.Stride += count // uint8 is 1 byte
	return vl
}

// Apply applies the layout to a vertex array
func (vl *VertexLayout) Apply(va *VertexArray) {
	// Update stride for all attributes
	for i := range vl.Attributes {
		vl.Attributes[i].Stride = vl.Stride
	}

	// Add all attributes to the vertex array
	for _, attr := range vl.Attributes {
		va.AddAttribute(attr)
	}
}

// Mesh represents a complete mesh with vertex and index data
type Mesh struct {
	VAO *VertexArray
	VBO *VertexBuffer
	IBO *IndexBuffer
}

// NewMesh creates a new mesh
func NewMesh(vertices []float32, indices []uint32, layout *VertexLayout) (*Mesh, error) {
	// Create vertex buffer
	vbo, err := NewVertexBuffer(vertices, StaticDraw)
	if err != nil {
		return nil, fmt.Errorf("failed to create vertex buffer: %w", err)
	}

	// Create index buffer (optional)
	var ibo *IndexBuffer
	if len(indices) > 0 {
		ibo, err = NewIndexBuffer(indices, StaticDraw)
		if err != nil {
			vbo.Delete()
			return nil, fmt.Errorf("failed to create index buffer: %w", err)
		}
	}

	// Create vertex array
	vao, err := NewVertexArray()
	if err != nil {
		vbo.Delete()
		if ibo != nil {
			ibo.Delete()
		}
		return nil, fmt.Errorf("failed to create vertex array: %w", err)
	}

	// Set buffers
	vao.SetVertexBuffer(vbo)
	if ibo != nil {
		vao.SetIndexBuffer(ibo)
	}

	// Apply layout
	if layout != nil {
		layout.Apply(vao)
	}

	return &Mesh{
		VAO: vao,
		VBO: vbo,
		IBO: ibo,
	}, nil
}

// Draw draws the mesh
func (m *Mesh) Draw(mode uint32) {
	if m.IBO != nil {
		m.VAO.DrawIndexed(mode)
	} else {
		// Calculate vertex count from VBO size and assuming float32 vertices
		// This is a simplified approach - in practice you'd track vertex count
		vertexCount := m.VBO.Size / 4 / 3 // Assuming 3 floats per vertex
		m.VAO.Draw(mode, int32(vertexCount), 0)
	}
}

// Delete deletes all mesh resources
func (m *Mesh) Delete() {
	if m.VAO != nil {
		m.VAO.Delete()
	}
	if m.VBO != nil {
		m.VBO.Delete()
	}
	if m.IBO != nil {
		m.IBO.Delete()
	}
}
package resource_test

import (
	"os"
	"testing"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/yossideutsch/gogl/pkg/resource"
)

var testWindow *glfw.Window

func TestMain(m *testing.M) {
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		panic("Failed to initialize GLFW: " + err.Error())
	}
	defer glfw.Terminate()

	// Configure OpenGL context
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Visible, glfw.False)

	// Create window
	var err error
	testWindow, err = glfw.CreateWindow(100, 100, "Test", nil, nil)
	if err != nil {
		panic("Failed to create test window: " + err.Error())
	}
	defer testWindow.Destroy()

	// Make context current
	testWindow.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		panic("Failed to initialize OpenGL: " + err.Error())
	}

	// Run tests
	os.Exit(m.Run())
}

func TestVertexBufferCreation(t *testing.T) {
	data := []float32{1.0, 2.0, 3.0, 4.0}
	
	vbo, err := resource.NewVertexBuffer(data, resource.StaticDraw)
	if err != nil {
		t.Fatal("Failed to create vertex buffer:", err)
	}
	defer vbo.Delete()

	if vbo.ID == 0 {
		t.Error("Buffer ID should not be 0")
	}

	if vbo.Size != len(data)*4 {
		t.Error("Buffer size mismatch")
	}

	if vbo.Target != resource.ArrayBuffer {
		t.Error("Buffer target should be ArrayBuffer")
	}
}

func TestIndexBufferCreation(t *testing.T) {
	data := []uint32{0, 1, 2, 3, 4, 5}
	
	ibo, err := resource.NewIndexBuffer(data, resource.StaticDraw)
	if err != nil {
		t.Fatal("Failed to create index buffer:", err)
	}
	defer ibo.Delete()

	if ibo.ID == 0 {
		t.Error("Buffer ID should not be 0")
	}

	if ibo.Count != len(data) {
		t.Error("Index count mismatch")
	}

	if ibo.IndexType != gl.UNSIGNED_INT {
		t.Error("Index type should be UNSIGNED_INT")
	}
}

func TestIndexBuffer16Creation(t *testing.T) {
	data := []uint16{0, 1, 2, 3}
	
	ibo, err := resource.NewIndexBuffer16(data, resource.StaticDraw)
	if err != nil {
		t.Fatal("Failed to create 16-bit index buffer:", err)
	}
	defer ibo.Delete()

	if ibo.Count != len(data) {
		t.Error("Index count mismatch")
	}

	if ibo.IndexType != gl.UNSIGNED_SHORT {
		t.Error("Index type should be UNSIGNED_SHORT")
	}
}

func TestUniformBufferCreation(t *testing.T) {
	size := 256 // UBO size
	
	ubo, err := resource.NewUniformBuffer(size, resource.DynamicDraw)
	if err != nil {
		t.Fatal("Failed to create uniform buffer:", err)
	}
	defer ubo.Delete()

	if ubo.ID == 0 {
		t.Error("Buffer ID should not be 0")
	}

	if ubo.Size != size {
		t.Error("Buffer size mismatch")
	}

	if ubo.Target != resource.UniformBufferTarget {
		t.Error("Buffer target should be UniformBufferTarget")
	}
}

func TestVertexArrayCreation(t *testing.T) {
	vao, err := resource.NewVertexArray()
	if err != nil {
		t.Fatal("Failed to create vertex array:", err)
	}
	defer vao.Delete()

	if vao.ID == 0 {
		t.Error("VAO ID should not be 0")
	}

	if len(vao.Attributes) != 0 {
		t.Error("New VAO should have no attributes")
	}
}

func TestVertexLayout(t *testing.T) {
	layout := resource.NewVertexLayout().
		AddFloat(0, 3).  // Position
		AddFloat(1, 2).  // UV
		AddInt(2, 1)     // ID

	if len(layout.Attributes) != 3 {
		t.Error("Layout should have 3 attributes")
	}

	expectedStride := int32(3*4 + 2*4 + 1*4) // 3 floats + 2 floats + 1 int
	if layout.Stride != expectedStride {
		t.Errorf("Expected stride %d, got %d", expectedStride, layout.Stride)
	}

	// Check first attribute (position)
	pos := layout.Attributes[0]
	if pos.Location != 0 || pos.Size != 3 || pos.Type != resource.Float {
		t.Error("Position attribute configuration incorrect")
	}

	// Check second attribute (UV)
	uv := layout.Attributes[1]
	if uv.Location != 1 || uv.Size != 2 || uv.Offset != 12 { // 3 floats * 4 bytes
		t.Error("UV attribute configuration incorrect")
	}
}

func TestMeshCreation(t *testing.T) {
	vertices := []float32{
		// Triangle
		-0.5, -0.5, 0.0,
		 0.5, -0.5, 0.0,
		 0.0,  0.5, 0.0,
	}

	indices := []uint32{0, 1, 2}

	layout := resource.NewVertexLayout().AddFloat(0, 3)

	mesh, err := resource.NewMesh(vertices, indices, layout)
	if err != nil {
		t.Fatal("Failed to create mesh:", err)
	}
	defer mesh.Delete()

	if mesh.VAO == nil {
		t.Error("Mesh should have a VAO")
	}

	if mesh.VBO == nil {
		t.Error("Mesh should have a VBO")
	}

	if mesh.IBO == nil {
		t.Error("Mesh should have an IBO")
	}

	if mesh.IBO.Count != 3 {
		t.Error("IBO should have 3 indices")
	}
}

func TestMeshWithoutIndices(t *testing.T) {
	vertices := []float32{
		-0.5, -0.5, 0.0,
		 0.5, -0.5, 0.0,
		 0.0,  0.5, 0.0,
	}

	layout := resource.NewVertexLayout().AddFloat(0, 3)

	mesh, err := resource.NewMesh(vertices, nil, layout)
	if err != nil {
		t.Fatal("Failed to create mesh without indices:", err)
	}
	defer mesh.Delete()

	if mesh.IBO != nil {
		t.Error("Mesh without indices should not have an IBO")
	}
}

func TestTexture2DCreation(t *testing.T) {
	config := resource.DefaultTextureConfig()
	
	texture, err := resource.NewTexture2D(256, 256, resource.FormatRGBA, config)
	if err != nil {
		t.Fatal("Failed to create texture:", err)
	}
	defer texture.Delete()

	if texture.ID == 0 {
		t.Error("Texture ID should not be 0")
	}

	if texture.Width != 256 || texture.Height != 256 {
		t.Error("Texture dimensions incorrect")
	}

	if texture.Format != resource.FormatRGBA {
		t.Error("Texture format incorrect")
	}
}

func TestTextureArrayCreation(t *testing.T) {
	config := resource.DefaultTextureConfig()
	
	texArray, err := resource.NewTextureArray(128, 128, 4, resource.FormatRGBA, config)
	if err != nil {
		t.Fatal("Failed to create texture array:", err)
	}
	defer texArray.Delete()

	if texArray.ID == 0 {
		t.Error("Texture array ID should not be 0")
	}

	if texArray.Layers != 4 {
		t.Error("Texture array should have 4 layers")
	}
}

func TestTextureManager(t *testing.T) {
	tm := resource.NewTextureManager()

	// Since we don't have actual image files in tests,
	// we'll just test the manager structure
	if tm == nil {
		t.Fatal("Failed to create texture manager")
	}

	// Test getting non-existent texture
	tex := tm.Get("nonexistent")
	if tex != nil {
		t.Error("Should return nil for non-existent texture")
	}
}

func TestBufferPool(t *testing.T) {
	pool := resource.NewBufferPool()

	// Acquire a buffer
	buf1, err := pool.Acquire(resource.ArrayBuffer, 1024, resource.StaticDraw)
	if err != nil {
		t.Fatal("Failed to acquire buffer:", err)
	}

	if buf1.ID == 0 {
		t.Error("Buffer ID should not be 0")
	}

	// Acquire another buffer
	buf2, err := pool.Acquire(resource.ArrayBuffer, 512, resource.StaticDraw)
	if err != nil {
		t.Fatal("Failed to acquire second buffer:", err)
	}

	// Release first buffer
	pool.Release(buf1)

	// Acquire buffer again - should reuse the released one
	buf3, err := pool.Acquire(resource.ArrayBuffer, 500, resource.StaticDraw)
	if err != nil {
		t.Fatal("Failed to acquire reused buffer:", err)
	}

	// Should reuse buf1 since it's large enough
	if buf3.ID != buf1.ID {
		t.Error("Pool should reuse released buffer")
	}

	// Clean up
	pool.Release(buf2)
	pool.Release(buf3)
	pool.Clear()
}
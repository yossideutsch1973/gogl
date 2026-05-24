// Showcase demo: a fullscreen raymarched scene rendered with the gogl
// shader/pipeline packages. All geometry lives in the fragment shader; the
// host program just drives a fullscreen triangle.
//
// Run interactively:
//
//	go run ./cmd/examples/showcase
//
// Render a single off-screen frame and write it to PNG (used to produce
// the README screenshot):
//
//	go run ./cmd/examples/showcase -screenshot docs/showcase.png \
//	    -width 1280 -height 720 -time 12.0
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/yossideutsch1973/gogl/pkg/pipeline"
	"github.com/yossideutsch1973/gogl/pkg/shader"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	var (
		screenshot string
		width      int
		height     int
		fixedTime  float64
		vertPath   string
		fragPath   string
	)
	flag.StringVar(&screenshot, "screenshot", "", "render one frame off-screen, write PNG here, and exit")
	flag.IntVar(&width, "width", 1280, "window/framebuffer width")
	flag.IntVar(&height, "height", 720, "window/framebuffer height")
	flag.Float64Var(&fixedTime, "time", 7.5, "uTime value to use for -screenshot mode")
	flag.StringVar(&vertPath, "vert", "shaders/vertex/fullscreen_triangle.vert", "vertex shader path")
	flag.StringVar(&fragPath, "frag", "shaders/fragment/raymarch_showcase.frag", "fragment shader path")
	flag.Parse()

	if err := glfw.Init(); err != nil {
		log.Fatalf("init glfw: %v", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	if screenshot != "" {
		glfw.WindowHint(glfw.Visible, glfw.False)
	}

	win, err := glfw.CreateWindow(width, height, "GoGL Showcase", nil, nil)
	if err != nil {
		log.Fatalf("create window: %v", err)
	}
	win.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatalf("init gl: %v", err)
	}

	prog, vao := buildPipeline(vertPath, fragPath)
	defer prog.Delete()
	defer gl.DeleteVertexArrays(1, &vao)

	resLoc := prog.GetUniformLocation("uResolution")
	timeLoc := prog.GetUniformLocation("uTime")

	p := pipeline.New()
	if err := p.SetState(pipeline.NewBuilder().
		WithProgram(prog).
		WithDepthTest(false, false, pipeline.DepthAlways).
		WithCulling(false, pipeline.CullNone).
		WithViewport(0, 0, int32(width), int32(height)).
		Build()); err != nil {
		log.Fatalf("set pipeline state: %v", err)
	}

	if screenshot != "" {
		renderOneFrame(prog, vao, resLoc, timeLoc, float32(width), float32(height), float32(fixedTime))
		if err := savePNG(screenshot, width, height); err != nil {
			log.Fatalf("write screenshot: %v", err)
		}
		fmt.Printf("wrote %s (%dx%d)\n", screenshot, width, height)
		return
	}

	runLoop(win, prog, vao, resLoc, timeLoc)
}

func buildPipeline(vertPath, fragPath string) (*shader.Program, uint32) {
	vert, err := shader.CompileShaderFromFile(vertPath, shader.VertexShader)
	if err != nil {
		log.Fatalf("compile vertex shader %s: %v", vertPath, err)
	}
	frag, err := shader.CompileShaderFromFile(fragPath, shader.FragmentShader)
	if err != nil {
		log.Fatalf("compile fragment shader %s: %v", fragPath, err)
	}
	prog, err := shader.CreateProgram(vert, frag)
	if err != nil {
		log.Fatalf("link program: %v", err)
	}

	// We draw a single fullscreen triangle generated entirely from
	// gl_VertexID — no vertex buffer required. Core profile still
	// requires a bound VAO for the draw call to be valid.
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	return prog, vao
}

func renderOneFrame(prog *shader.Program, vao uint32, resLoc, timeLoc int32, w, h, t float32) {
	prog.Use()
	_ = prog.SetUniform1f(timeLoc, t)
	gl.Uniform2f(resLoc, w, h)

	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
	gl.BindVertexArray(0)
	gl.Finish()
}

func runLoop(win *glfw.Window, prog *shader.Program, vao uint32, resLoc, timeLoc int32) {
	start := time.Now()
	for !win.ShouldClose() {
		if win.GetKey(glfw.KeyEscape) == glfw.Press {
			win.SetShouldClose(true)
		}

		fw, fh := win.GetFramebufferSize()
		gl.Viewport(0, 0, int32(fw), int32(fh))

		prog.Use()
		_ = prog.SetUniform1f(timeLoc, float32(time.Since(start).Seconds()))
		gl.Uniform2f(resLoc, float32(fw), float32(fh))

		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
		gl.BindVertexArray(0)

		win.SwapBuffers()
		glfw.PollEvents()
	}
}

func savePNG(path string, w, h int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	pixels := make([]uint8, w*h*4)
	gl.PixelStorei(gl.PACK_ALIGNMENT, 1)
	gl.ReadPixels(0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels[0]))

	// OpenGL origin is bottom-left; PNG is top-left. Flip rows in place.
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	stride := w * 4
	for y := 0; y < h; y++ {
		srcRow := pixels[y*stride : (y+1)*stride]
		dstY := h - 1 - y
		copy(img.Pix[dstY*img.Stride:dstY*img.Stride+stride], srcRow)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

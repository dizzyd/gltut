package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"io/ioutil"

	"image"

	_ "image/jpeg"
	_ "image/png"

	"image/draw"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var t1 = []float32{
	// Positions   Colors    Texture Coords
	0.5, 0.5, 0.0, 1.0, 0.0, 0.0, 1.0, 1.0,
	0.5, -0.5, 0.0, 0.0, 1.0, 0.0, 1.0, 0.0,
	-0.5, -0.5, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0,
	-0.5, 0.5, 0.0, 1.0, 1.0, 0.0, 0.0, 1.0,
}

var t1Indices = []uint32{
	0, 1, 3,
	1, 2, 3,
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}
}

func init() {
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.False)

	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	fmt.Printf("%s %s\n", gl.GoStr(gl.GetString(gl.RENDERER)), gl.GoStr(gl.GetString(gl.VERSION)))

	window.SetKeyCallback(keyCallback)

	_, vao1 := initTriangle(t1, t1Indices)

	// Load up a program
	p1, err := compileProgram("shaders/vert3.glsl", "shaders/frag3.glsl")
	if err != nil {
		panic(err)
	}

	gl.UseProgram(p1)

	texture1, err := loadTexture("textures/container.jpg")
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

	texture2, err := loadTexture("textures/awesomeface.png")
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture1)
	gl.Uniform1i(gl.GetUniformLocation(p1, gl.Str("texture1\x00")), 0)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, texture2)
	gl.Uniform1i(gl.GetUniformLocation(p1, gl.Str("texture2\x00")), 1)

	horizOffsetStr := gl.Str("horizOffset\x00")
	horizOffsetLoc := gl.GetUniformLocation(p1, horizOffsetStr)

	for !window.ShouldClose() {
		glfw.PollEvents()

		timeValue := glfw.GetTime()
		horizOffset := float32((math.Sin(timeValue) / 2) + 0.5)
		gl.Uniform1f(horizOffsetLoc, horizOffset)

		gl.ClearColor(0.3, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.BindVertexArray(vao1)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)

		gl.BindVertexArray(0)

		window.SwapBuffers()
	}

}

func initTriangle(vertices []float32, indices []uint32) (uint32, uint32) {
	// Setup the VBO/VAO for t1
	var vbo, vao, ebo uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Positions
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*4, nil)
	gl.EnableVertexAttribArray(0)

	// Colors
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 8*4, unsafe.Pointer(uintptr(3*4)))
	gl.EnableVertexAttribArray(1)

	// Texture Coords
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 8*4, unsafe.Pointer(uintptr(6*4)))
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)
	return vbo, vao
}

func compileProgram(vertexShaderName string, fragmentShaderName string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderName, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	defer gl.DeleteShader(vertexShader)

	fragShader, err := compileShader(fragmentShaderName, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}
	defer gl.DeleteShader(fragShader)

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status != gl.TRUE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength)+1)
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("Failed to compile program %s / %s: %v", vertexShaderName, fragmentShaderName, log)
	}

	return program, nil
}

func compileShader(sourceFilename string, shaderType uint32) (uint32, error) {
	// Load raw bytes from the source file
	shaderBytes, err := ioutil.ReadFile(sourceFilename)
	if err != nil {
		return 0, err
	}

	// Convert raw bytes to format suitable for loading into OpenGL
	shaderBytesLen := int32(len(shaderBytes))
	shaderStr, shaderStrFree := gl.Strs(string(shaderBytes))

	// Initialize a shader
	shader := gl.CreateShader(shaderType)
	gl.ShaderSource(shader, 1, shaderStr, &shaderBytesLen)
	shaderStrFree()
	gl.CompileShader(shader)

	// Check for errors
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status != gl.TRUE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength)+1)
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("Failed to compile shader %s: %v", sourceFilename, log)
	}

	fmt.Printf("Loaded shader: %s\n", sourceFilename)
	return shader, nil
}

func loadImage(filename string) (*image.RGBA, error) {
	// Load the image from disk
	imFile, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Failed to open image %s: %+v\n", filename, err)
	}
	defer imFile.Close()

	im, filetype, err := image.Decode(imFile)
	if err != nil {
		return nil, fmt.Errorf("invalid image: %s: %+v", filename, err)
	}
	fmt.Printf("Loaded %s as %s!\n", filename, filetype)
	switch actualIm := im.(type) {
	case *image.RGBA:
		return actualIm, nil
	default:
		imCopy := image.NewRGBA(actualIm.Bounds())
		draw.Draw(imCopy, actualIm.Bounds(), actualIm, image.Pt(0, 0), draw.Src)
		return imCopy, nil
	}
}

func loadTexture(filename string) (uint32, error) {
	img, err := loadImage(filename)
	if err != nil {
		return 0, err
	}

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(img.Bounds().Dx()), int32(img.Bounds().Dy()),
		0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return texture, nil
}

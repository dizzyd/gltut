package main

import (
	"fmt"
	"runtime"
	"strings"

	"io/ioutil"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var t1 = []float32{
	-0.25, -0.25, 0.0,
	0.25, -0.25, 0.0,
	0.0, 0.25, 0.0,
}

var t2 = []float32{
	0.25, 0.5, 0.0,
	-0.25, 0.5, 0.0,
	0.0, 0.25, 0.0,
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

	_, vao1 := initTriangle(t1)
	_, vao2 := initTriangle(t2)

	// Load up a program
	p1, err := compileProgram("shaders/vert1.glsl", "shaders/frag1.glsl")
	if err != nil {
		panic(err)
	}

	p2, err := compileProgram("shaders/vert1.glsl", "shaders/frag1a.glsl")
	if err != nil {
		panic(err)
	}

	for !window.ShouldClose() {
		glfw.PollEvents()

		gl.ClearColor(0.3, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.UseProgram(p1)

		gl.BindVertexArray(vao1)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)

		gl.UseProgram(p2)

		gl.BindVertexArray(vao2)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)

		gl.BindVertexArray(0)

		window.SwapBuffers()
	}
}

func initTriangle(vertices []float32) (uint32, uint32) {
	// Setup the VBO/VAO for t1
	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 12, nil)
	gl.EnableVertexAttribArray(0)
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

	return shader, nil
}

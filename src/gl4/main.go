package main

import (
	"fmt"
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
	"github.com/go-gl/mathgl/mgl32"
)

var t1 = []float32{
	-0.5, -0.5, -0.5, 0.0, 0.0,
	0.5, -0.5, -0.5, 1.0, 0.0,
	0.5, 0.5, -0.5, 1.0, 1.0,
	0.5, 0.5, -0.5, 1.0, 1.0,
	-0.5, 0.5, -0.5, 0.0, 1.0,
	-0.5, -0.5, -0.5, 0.0, 0.0,

	-0.5, -0.5, 0.5, 0.0, 0.0,
	0.5, -0.5, 0.5, 1.0, 0.0,
	0.5, 0.5, 0.5, 1.0, 1.0,
	0.5, 0.5, 0.5, 1.0, 1.0,
	-0.5, 0.5, 0.5, 0.0, 1.0,
	-0.5, -0.5, 0.5, 0.0, 0.0,

	-0.5, 0.5, 0.5, 1.0, 0.0,
	-0.5, 0.5, -0.5, 1.0, 1.0,
	-0.5, -0.5, -0.5, 0.0, 1.0,
	-0.5, -0.5, -0.5, 0.0, 1.0,
	-0.5, -0.5, 0.5, 0.0, 0.0,
	-0.5, 0.5, 0.5, 1.0, 0.0,

	0.5, 0.5, 0.5, 1.0, 0.0,
	0.5, 0.5, -0.5, 1.0, 1.0,
	0.5, -0.5, -0.5, 0.0, 1.0,
	0.5, -0.5, -0.5, 0.0, 1.0,
	0.5, -0.5, 0.5, 0.0, 0.0,
	0.5, 0.5, 0.5, 1.0, 0.0,

	-0.5, -0.5, -0.5, 0.0, 1.0,
	0.5, -0.5, -0.5, 1.0, 1.0,
	0.5, -0.5, 0.5, 1.0, 0.0,
	0.5, -0.5, 0.5, 1.0, 0.0,
	-0.5, -0.5, 0.5, 0.0, 0.0,
	-0.5, -0.5, -0.5, 0.0, 1.0,

	-0.5, 0.5, -0.5, 0.0, 1.0,
	0.5, 0.5, -0.5, 1.0, 1.0,
	0.5, 0.5, 0.5, 1.0, 0.0,
	0.5, 0.5, 0.5, 1.0, 0.0,
	-0.5, 0.5, 0.5, 0.0, 0.0,
	-0.5, 0.5, -0.5, 0.0, 1.0,
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

	vao1 := initTriangle(t1)

	// Load up a program
	p1, err := compileProgram("shaders/vert4.glsl", "shaders/frag4.glsl")
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

	gl.Enable(gl.DEPTH_TEST)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture1)
	gl.Uniform1i(gl.GetUniformLocation(p1, gl.Str("texture1\x00")), 0)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, texture2)
	gl.Uniform1i(gl.GetUniformLocation(p1, gl.Str("texture2\x00")), 1)

	// Rotate vertices around X-axis -55 degrees
	//model := mgl32.HomogRotate3D(mgl32.DegToRad(-55.0), mgl32.Vec3{1.0, 0.0, 0.0})
	// Step back -3
	view := mgl32.Translate3D(0.0, 0.0, -3.0)
	// Project??
	projection := mgl32.Perspective(45.0, 640/480, 0.1, 100.0)

	modelLoc := gl.GetUniformLocation(p1, gl.Str("model\x00"))
	viewLoc := gl.GetUniformLocation(p1, gl.Str("view\x00"))
	projLoc := gl.GetUniformLocation(p1, gl.Str("projection\x00"))

	gl.UniformMatrix4fv(viewLoc, 1, false, (*float32)(unsafe.Pointer(&view[0])))
	gl.UniformMatrix4fv(projLoc, 1, false, (*float32)(unsafe.Pointer(&projection[0])))

	for !window.ShouldClose() {
		glfw.PollEvents()

		t := float32(glfw.GetTime())
		model := mgl32.HomogRotate3D(mgl32.DegToRad(t*50.0), mgl32.Vec3{0.5, 1.0, 0.0})
		gl.UniformMatrix4fv(modelLoc, 1, false, (*float32)(unsafe.Pointer(&model[0])))

		// transform = transform.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75))
		// gl.UniformMatrix4fv(transformLoc, 1, false, (*float32)(unsafe.Pointer(&transform[0])))
		// transformLoc := gl.GetUniformLocation(p1, gl.Str("transform\x00"))

		gl.ClearColor(0.3, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.BindVertexArray(vao1)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)

		gl.BindVertexArray(0)

		window.SwapBuffers()
	}

}

func initTriangle(vertices []float32) uint32 {
	// Setup the VBO/VAO
	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Positions
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
	gl.EnableVertexAttribArray(0)

	// Texture Coords
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, unsafe.Pointer(uintptr(3*4)))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
	return vao
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

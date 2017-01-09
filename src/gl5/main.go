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

var cubes = []mgl32.Vec3{
	mgl32.Vec3{0.0, 0.0, 0.0},
	mgl32.Vec3{2.0, 5.0, -15.0},
	mgl32.Vec3{1.5, -2.2, -2.5},
	mgl32.Vec3{3.8, -2.0, -12.3},
	mgl32.Vec3{2.4, -0.4, -3.5},
	mgl32.Vec3{1.7, 3.0, -7.5},
	mgl32.Vec3{1.3, -2.0, -2.5},
	mgl32.Vec3{1.5, 2.0, -2.5},
	mgl32.Vec3{1.5, 0.2, -1.5},
	mgl32.Vec3{1.3, 1.0, -1.5},
}

var camera = newCamera()

const gWidth = 800
const gHeight = 600

var lastX = gWidth / 2.0
var lastY = gHeight / 2.0
var firstMouse = true

var keys [1024]bool

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	} else {
		keys[key] = (action == glfw.Press || action == glfw.Repeat)
	}
}

func mouseCallback(w *glfw.Window, x, y float64) {
	if firstMouse {
		lastX = x
		lastY = y
		firstMouse = false
	}

	xoffset := x - lastX
	yoffset := lastY - y

	lastX = x
	lastY = y

	camera.processMousePos(xoffset, yoffset)
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

	window, err := glfw.CreateWindow(gWidth, gHeight, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}

	window.SetKeyCallback(keyCallback)
	window.SetCursorPosCallback(mouseCallback)
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	fmt.Printf("%s %s\n", gl.GoStr(gl.GetString(gl.RENDERER)), gl.GoStr(gl.GetString(gl.VERSION)))

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

	projection := mgl32.Perspective(45.0, gWidth/gHeight, 0.1, 100.0)

	modelLoc := gl.GetUniformLocation(p1, gl.Str("model\x00"))
	viewLoc := gl.GetUniformLocation(p1, gl.Str("view\x00"))
	projLoc := gl.GetUniformLocation(p1, gl.Str("projection\x00"))

	gl.UniformMatrix4fv(projLoc, 1, false, (*float32)(unsafe.Pointer(&projection[0])))

	lastTime := glfw.GetTime()

	for !window.ShouldClose() {
		glfw.PollEvents()

		currTime := glfw.GetTime()
		doMovement(float32(currTime - lastTime))
		lastTime = currTime

		view := camera.viewMatrix()
		gl.UniformMatrix4fv(viewLoc, 1, false, (*float32)(unsafe.Pointer(&view[0])))

		gl.ClearColor(0.3, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.BindVertexArray(vao1)

		for _, cube := range cubes {
			model := mgl32.Translate3D(cube[0], cube[1], cube[2])
			gl.UniformMatrix4fv(modelLoc, 1, false, (*float32)(unsafe.Pointer(&model[0])))
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		}

		gl.BindVertexArray(0)

		window.SwapBuffers()
	}

}

func doMovement(deltaTime float32) {
	if keys[glfw.KeyW] {
		camera.processKeyboard(MoveForward, deltaTime)
	}
	if keys[glfw.KeyS] {
		camera.processKeyboard(MoveBackward, deltaTime)
	}
	if keys[glfw.KeyA] {
		camera.processKeyboard(MoveLeft, deltaTime)
	}
	if keys[glfw.KeyD] {
		camera.processKeyboard(MoveRight, deltaTime)
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

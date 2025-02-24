package main

import (
	"GoLas/readLas"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"strings"

	// "GoGIS/gospatial"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	width          = 1000
	height         = 1000
	yaw    float64 = 0.0
	pitch  float64 = 0.0
	roll   float64 = 0.0
	// lastX       float64 = float64(*width) / 2
	// lastY       float64 = float64(*height) / 2
	firstMouse          = true
	Tx          float32 = 0
	Ty          float32 = 0
	Tz          float32 = 0
	ScaleXY     float32 = 1
	Tsens       float32 = 0.01
	sensitivity float64 = 0.1
)

const (
	vertexShaderSource = `
	#version 410
	layout (location = 0) in vec3 aPos;   // the position variable has attribute position 0
	layout (location = 1) in vec3 aColor; // the color variable has attribute position 1
	uniform mat4 model;
	uniform mat4 projection;
	out vec3 ourColor; // output a color to the fragment shader
	void main()
	{
		gl_PointSize = 2.0;
		gl_Position = model * vec4(aPos, 1.0);
		ourColor = aColor; // set ourColor to the input color we got from the vertex data
	}

	` + "\x00"

	fragmentShaderSource = `
	#version 410
	out vec4 FragColor;
	in vec3 ourColor;
	float near = 0.1;
	float far  = 100.0;

	float LinearizeDepth(float depth)
	{
		float z = depth * 2.0 - 1.0; // back to NDC
		return (2.0 * near * far) / (far + near - z * (far - near));
	}

	void main()
	{
		float depth = LinearizeDepth(gl_FragCoord.z) / far; // Normalize depth
		vec3 depthColor = vec3(depth); // Convert depth to grayscale
		FragColor = vec4(mix(ourColor, depthColor, 0.0), 1.0); // Blend color with depth
	}

	` + "\x00"
	// vertexShaderSource = `
	// #version 410
	// layout(location = 0) in vec3 vp; // Vertex Position

	// out float fragZ; // Send Z to fragment shader
	// uniform mat4 model;

	// void main() {
	// 	fragZ = vp.z; // Pass the Z coordinate
	// 	gl_Position = model * vec4(vp, 1.0);
	// }

	// ` + "\x00"

	// fragmentShaderSource = `
	// #version 410
	// in float fragZ; // Received from vertex shader
	// out vec4 frag_colour;

	// void main() {
	// 	float colorValue = (fragZ); // Normalize Z from [-1,1] to [0,1]
	// 	frag_colour = vec4(1.0-colorValue , colorValue , 0.0 , 1.0); // Gradient from Blue to Red
	// }
	// ` + "\x00"
)

var (
	triangle = []float32{}
	line     = []float32{}

	vertex = []float32{}
)

func main() {

	filePath := os.Args
	var fileData readLas.LAS = readLas.Read_file(filePath[1])
	fmt.Println(fileData.Info)

	Hratio := (fileData.Info.Extent.MaxY - fileData.Info.Extent.MinY) / (fileData.Info.Extent.MaxX - fileData.Info.Extent.MinX)
	Vratio := (fileData.Info.Extent.MaxZ - fileData.Info.Extent.MinZ) / math.Sqrt(math.Pow(fileData.Info.Extent.MaxX-fileData.Info.Extent.MinX, 2)+math.Pow(fileData.Info.Extent.MaxY-fileData.Info.Extent.MinY, 2))

	fmt.Println(width, height)

	Xc := (fileData.Info.Extent.MaxX + fileData.Info.Extent.MinX) / 2
	Yc := (fileData.Info.Extent.MaxY + fileData.Info.Extent.MinY) / 2
	Zc := (fileData.Info.Extent.MaxZ + fileData.Info.Extent.MinZ) / 2
	for i := 0; i < len(fileData.Points); i++ {
		vertex = append(vertex, float32((float64(fileData.Points[i].X)*fileData.Info.Scale.X+float64(fileData.Info.Offset.X))-Xc)/float32(fileData.Info.Extent.MaxX-Xc))
		vertex = append(vertex, float32((float64(fileData.Points[i].Y)*fileData.Info.Scale.Y+float64(fileData.Info.Offset.Y))-Yc)/float32(fileData.Info.Extent.MaxY-Yc)*float32(Hratio))
		vertex = append(vertex, float32((float64(fileData.Points[i].Z)*fileData.Info.Scale.Z+float64(fileData.Info.Offset.Z))-Zc)/float32(fileData.Info.Extent.MinZ-Zc)*float32(Vratio))
		vertex = append(vertex, float32(fileData.Points[i].Red)/(float32(65535)))
		vertex = append(vertex, float32(fileData.Points[i].Green)/(float32(65535)))
		vertex = append(vertex, float32(fileData.Points[i].Blue)/(float32(65535)))
	}

	fileData.Points = []readLas.Point{}

	runtime.LockOSThread()

	window := initGlfw(width, width)
	window.MakeContextCurrent()
	// window.SetCursorPosCallback(mouseCallback)

	defer glfw.Terminate()
	program := initOpenGL()
	glfw.WindowHint(glfw.DepthBits, 24)
	gl.ClearColor(0.5, 0.5, 0.5, 1.0)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.DepthMask(true)
	gl.ClearDepth(1.0)

	// var left float32 = -2.0
	// var right float32 = 2.0
	// var bottom float32 = -1.5
	// var top float32 = 1.5
	// var near float32 = 0.1
	// var far float32 = 1000.0

	// projection := mgl32.Ortho(left, right, bottom, top, near, far)
	axisArr := makeVao(vertex, 0)
	for !window.ShouldClose() {
		keyboardInput(window)

		gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
		gl.ClearDepth(1.0)
		// projection := mgl32.Ortho(left, right, bottom, top, near, far)
		// aspect := float32(width) / float32(height)
		// projection := mgl32.Perspective(mgl32.DegToRad(45.0), aspect, 0.1, 1000.0)

		model := mgl32.Translate3D(Tx, Ty, Tz).Mul4(
			mgl32.HomogRotate3D(mgl32.DegToRad(float32(yaw)), mgl32.Vec3{0, 1, 0}),
		).Mul4(
			mgl32.HomogRotate3D(mgl32.DegToRad(float32(pitch)), mgl32.Vec3{1, 0, 0}),
		).Mul4(
			mgl32.HomogRotate3D(mgl32.DegToRad(float32(roll)), mgl32.Vec3{0, 0, 1}),
		).Mul4(mgl32.Scale3D(ScaleXY, ScaleXY, ScaleXY))

		modelLoc := gl.GetUniformLocation(program, gl.Str("model\x00"))
		gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

		// projLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
		// gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

		draw(axisArr, window, program, 2)

		glfw.PollEvents()
		window.SwapBuffers()
	}

}
func draw(vao uint32, window *glfw.Window, program uint32, size float32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gl.BindVertexArray(vao)

	// gl.PolygonMode(gl.DEPTH, gl.POINT)
	gl.PointSize(size)
	gl.DrawArrays(gl.POINTS, 0, int32(len(vertex)/6))

}

// func framebufferSizeCallback(window *glfw.Window, width int, height int) {
// 	gl.Viewport(0, 0, int32(width), int32(height))
// }

func initGlfw(width int, height int) *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Coba hehehe", nil, nil)
	if err != nil {
		panic(err)
	}
	// window.SetFramebufferSizeCallback(framebufferSizeCallback)

	window.MakeContextCurrent()

	return window
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog
}

// makeVao initializes and returns a vertex array from the points provided.
func makeVao(points []float32, i int32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(4*6), gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(4*6), gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}
func keyboardInput(window *glfw.Window) {
	// fmt.Println(yaw, pitch, roll)
	if glfw.Press == window.GetKey(glfw.KeyQ) {

		yaw += float64(Tsens) * 40
	} else if glfw.Press == window.GetKey(glfw.KeyE) {
		yaw -= float64(Tsens) * 40
	}
	if glfw.Press == window.GetKey(glfw.KeyW) {

		pitch += float64(Tsens) * 40
	} else if glfw.Press == window.GetKey(glfw.KeyS) {
		pitch -= float64(Tsens) * 40
	}
	if glfw.Press == window.GetKey(glfw.KeyA) {

		roll += float64(Tsens) * 40
	} else if glfw.Press == window.GetKey(glfw.KeyD) {
		roll -= float64(Tsens) * 40
	}
	// if glfw.Press == window.GetKey(glfw.KeySpace) {
	// 	pitch = 0
	// 	yaw = 0
	// 	roll = 0
	// 	Tx = 0
	// 	Ty = 0
	// 	Tz = 0
	// }
	if glfw.Press == window.GetKey(glfw.KeyUp) {
		Ty -= Tsens
	}
	if glfw.Press == window.GetKey(glfw.KeyDown) {
		Ty += Tsens
	}
	if glfw.Press == window.GetKey(glfw.KeyLeft) {
		Tx += Tsens
	}
	if glfw.Press == window.GetKey(glfw.KeyRight) {
		Tx -= Tsens
	}
	if glfw.Press == window.GetKey(glfw.KeySpace) {
		Tz += Tsens
	}
	if glfw.Press == window.GetKey(glfw.KeyLeftControl) {
		Tz -= Tsens
	}

	if glfw.Press == window.GetMouseButton(glfw.MouseButtonLeft) {
		ScaleXY += Tsens
	} else if glfw.Press == window.GetMouseButton(glfw.MouseButtonRight) {
		if ScaleXY > 0+Tsens {
			ScaleXY -= Tsens
		}
	}

}

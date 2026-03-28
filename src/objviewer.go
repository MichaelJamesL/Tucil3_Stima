package src

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 800
	screenHeight = 800
)

type Memory struct {
	vertices []Vec3
	faces    [][]int
	center   Vec3
	angleX   float64
	angleY   float64
	scale    float64
	buffer   *image.RGBA
	zBuffer  []float64
}

type screenVertex struct {
	x, y int
	z    float64
}

// Bresenham's Line Algorithm
func drawLine(x0, y0, x1, y1 int, img *image.RGBA, col color.RGBA) {
	dx := math.Abs(float64(x1 - x0))
	dy := math.Abs(float64(y1 - y0))
	sx, sy := -1, -1
	if x0 < x1 {
		sx = 1
	}
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		if x0 >= 0 && x0 < screenWidth && y0 >= 0 && y0 < screenHeight {
			offset := (y0*screenWidth + x0) * 4
			img.Pix[offset] = col.R
			img.Pix[offset+1] = col.G
			img.Pix[offset+2] = col.B
			img.Pix[offset+3] = col.A
		}

		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func putPixel(img *image.RGBA, x, y int, col color.RGBA) {
	if x < 0 || x >= screenWidth || y < 0 || y >= screenHeight {
		return
	}
	offset := (y*screenWidth + x) * 4
	img.Pix[offset] = col.R
	img.Pix[offset+1] = col.G
	img.Pix[offset+2] = col.B
	img.Pix[offset+3] = col.A
}

// Signed triangle edge test used for barycentric inside-triangle checks
func edgeFunction(ax, ay, bx, by, px, py float64) float64 {
	return (px-ax)*(by-ay) - (py-ay)*(bx-ax)
}

// Triangle rasterization with barycentric coordinates + z-buffer depth test
func drawFilledTriangle(v0, v1, v2 screenVertex, img *image.RGBA, zBuffer []float64, col color.RGBA) {
	minX := int(math.Max(0, math.Min(float64(v0.x), math.Min(float64(v1.x), float64(v2.x)))))
	maxX := int(math.Min(float64(screenWidth-1), math.Max(float64(v0.x), math.Max(float64(v1.x), float64(v2.x)))))
	minY := int(math.Max(0, math.Min(float64(v0.y), math.Min(float64(v1.y), float64(v2.y)))))
	maxY := int(math.Min(float64(screenHeight-1), math.Max(float64(v0.y), math.Max(float64(v1.y), float64(v2.y)))))

	ax, ay := float64(v0.x), float64(v0.y)
	bx, by := float64(v1.x), float64(v1.y)
	cx, cy := float64(v2.x), float64(v2.y)
	area := edgeFunction(ax, ay, bx, by, cx, cy)
	if area == 0 {
		return
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			px := float64(x) + 0.5
			py := float64(y) + 0.5

			w0 := edgeFunction(bx, by, cx, cy, px, py)
			w1 := edgeFunction(cx, cy, ax, ay, px, py)
			w2 := edgeFunction(ax, ay, bx, by, px, py)

			if (w0 >= 0 && w1 >= 0 && w2 >= 0 && area > 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0 && area < 0) {
				w0n := w0 / area
				w1n := w1 / area
				w2n := w2 / area
				z := w0n*v0.z + w1n*v1.z + w2n*v2.z

				idx := y*screenWidth + x
				if z > zBuffer[idx] {
					zBuffer[idx] = z
					putPixel(img, x, y, col)
				}
			}
		}
	}
}

// Update runs every frame to handle input
func (a *Memory) Update() error {
	// Rotation controls
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		a.angleY -= 0.05
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		a.angleY += 0.05
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		a.angleX -= 0.05
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		a.angleX += 0.05
	}

	// Zoom controls
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		a.scale *= 1.05
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		a.scale /= 1.05
	}

	return nil
}

// Draw steps:
// 1) model transform (center + yaw/pitch rotation)
// 2) orthographic projection
// 3) Lambert diffuse shading from face normal
// 4) triangle fill with z-buffer and optional wireframe overlay
func (a *Memory) Draw(screen *ebiten.Image) {
	// 1. Clear the screen (fill with black)
	for i := 0; i < len(a.buffer.Pix); i += 4 {
		a.buffer.Pix[i] = 0     // R
		a.buffer.Pix[i+1] = 0   // G
		a.buffer.Pix[i+2] = 0   // B
		a.buffer.Pix[i+3] = 255 // A
	}
	for i := range a.zBuffer {
		a.zBuffer[i] = math.Inf(-1)
	}

	// Put instructions on the screen
	instructions := "Controls: Arrow Keys = Rotate | W/S = Zoom"

	sinY, cosY := math.Sin(a.angleY), math.Cos(a.angleY)
	sinX, cosX := math.Sin(a.angleX), math.Cos(a.angleX)
	lightX, lightY, lightZ := 0.3, 0.6, 1.0
	lightLen := math.Sqrt(lightX*lightX + lightY*lightY + lightZ*lightZ)
	lightX, lightY, lightZ = lightX/lightLen, lightY/lightLen, lightZ/lightLen

	// 2. Process and draw each face
	for _, face := range a.faces {
		if len(face) < 3 {
			continue
		}

		transformed := make([]Vec3, 0, len(face))
		projected := make([]screenVertex, 0, len(face))

		for i := 0; i < len(face); i++ {
			v := a.vertices[face[i]]
			x := v.x - a.center.x
			y := v.y - a.center.y
			z := v.z - a.center.z

			// Apply Y-axis rotation (Yaw)
			x1 := x*cosY + z*sinY
			z1 := -x*sinY + z*cosY
			y1 := y

			// Apply X-axis rotation (Pitch)
			y2 := y1*cosX - z1*sinX
			z2 := y1*sinX + z1*cosX
			transformed = append(transformed, Vec3{x: x1, y: y2, z: z2})

			// Orthographic Projection & Screen scaling
			screenX := int((x1 * a.scale) + (screenWidth / 2))
			screenY := int((-y2 * a.scale) + (screenHeight / 2)) // Invert Y for screen coordinates

			projected = append(projected, screenVertex{x: screenX, y: screenY, z: z2})
		}

		// Face normal from cross product of two triangle edges
		nx := (transformed[1].y-transformed[0].y)*(transformed[2].z-transformed[0].z) - (transformed[1].z-transformed[0].z)*(transformed[2].y-transformed[0].y)
		ny := (transformed[1].z-transformed[0].z)*(transformed[2].x-transformed[0].x) - (transformed[1].x-transformed[0].x)*(transformed[2].z-transformed[0].z)
		nz := (transformed[1].x-transformed[0].x)*(transformed[2].y-transformed[0].y) - (transformed[1].y-transformed[0].y)*(transformed[2].x-transformed[0].x)
		nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
		if nLen == 0 {
			continue
		}
		nx, ny, nz = nx/nLen, ny/nLen, nz/nLen

		// Lambert intensity = dot(normal, lightDir)
		brightness := nx*lightX + ny*lightY + nz*lightZ
		if brightness < 0 {
			brightness = 0
		}
		c := uint8(30 + 225*brightness)
		fill := color.RGBA{c, c, c, 255}

		// Fan triangulation for polygonal faces from OBJ
		for i := 1; i < len(projected)-1; i++ {
			drawFilledTriangle(projected[0], projected[i], projected[i+1], a.buffer, a.zBuffer, fill)
		}

		edge := color.RGBA{220, 220, 220, 255}
		for i := 0; i < len(projected); i++ {
			p0 := projected[i]
			p1 := projected[(i+1)%len(projected)]
			drawLine(p0.x, p0.y, p1.x, p1.y, a.buffer, edge)
		}
	}

	// 3. Write our raw pixel buffer to the screen
	screen.WritePixels(a.buffer.Pix)
	ebitenutil.DebugPrintAt(screen, instructions, 10, 10)
}

func (a *Memory) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func ObjViewer() {
	var inputFile string
	fmt.Println("OBJ File Path:")
	fmt.Scanf("%s", &inputFile)

	app := &Memory{
		scale:  screenWidth / 2.5,
		buffer: image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight)),
		zBuffer: func() []float64 {
			z := make([]float64, screenWidth*screenHeight)
			for i := range z {
				z[i] = math.Inf(-1)
			}
			return z
		}(),
	}

	// Parse OBJ File
	file, err := os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}

		if parts[0] == "v" {
			x, _ := strconv.ParseFloat(parts[1], 64)
			y, _ := strconv.ParseFloat(parts[2], 64)
			z, _ := strconv.ParseFloat(parts[3], 64)
			app.vertices = append(app.vertices, Vec3{x, y, z})
		} else if parts[0] == "f" {
			var faceIndices []int
			for i := 1; i < len(parts); i++ {
				vData := strings.Split(parts[i], "/")
				idx, _ := strconv.Atoi(vData[0])
				faceIndices = append(faceIndices, idx-1)
			}
			app.faces = append(app.faces, faceIndices)
		}
	}

	if len(app.vertices) == 0 {
		panic("no vertices found in OBJ")
	}

	minV, maxV := getBoundingBox(app.vertices)

	app.center = Vec3{
		x: (minV.x + maxV.x) / 2,
		y: (minV.y + maxV.y) / 2,
		z: (minV.z + maxV.z) / 2,
	}

	span := math.Max(maxV.x-minV.x, math.Max(maxV.y-minV.y, maxV.z-minV.z))
	if span > 0 {
		app.scale = 0.7 * math.Min(float64(screenWidth), float64(screenHeight)) / span
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("OBJ Viewer")
	if err := ebiten.RunGame(app); err != nil {
		panic(err)
	}
}

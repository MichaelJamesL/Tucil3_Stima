package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Vec3 struct {
	x, y, z float64
}

type Cube struct {
	Min, Max Vec3
}

type Triangle struct {
	v1, v2, v3 Vec3
}

type Node struct {
	IsVoxel  bool
	Center   Vec3
	Children [8]*Node
}

func getBoundingBox(vertices []Vec3) (Vec3, Vec3) {
	min := Vec3{vertices[0].x, vertices[0].y, vertices[0].z}
	max := Vec3{vertices[0].x, vertices[0].y, vertices[0].z}

	for _, v := range vertices {
		if v.x < min.x {
			min.x = v.x
		}
		if v.y < min.y {
			min.y = v.y
		}
		if v.z < min.z {
			min.z = v.z
		}
		if v.x > max.x {
			max.x = v.x
		}
		if v.y > max.y {
			max.y = v.y
		}
		if v.z > max.z {
			max.z = v.z
		}
	}
	return min, max
}

func readOBJ(path string) ([]Vec3, []Triangle, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var vertices []Vec3
	var triangles []Triangle

	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			// Process each line to extract vertex and triangle data
			if line[0] == 'v' {
				var v Vec3
				fmt.Sscanf(line, "v %f %f %f", &v.x, &v.y, &v.z)
				vertices = append(vertices, v)
			} else if line[0] == 'f' {
				parts := strings.Fields(line)
				if len(parts) < 4 {
					continue
				}

				parseIndex := func(token string) int {
					slash := strings.Index(token, "/")
					if slash != -1 {
						token = token[:slash]
					}
					idx, err := strconv.Atoi(token)
					if err != nil {
						return 0
					}
					return idx
				}

				i := parseIndex(parts[1])
				j := parseIndex(parts[2])
				k := parseIndex(parts[3])
				if i <= 0 || j <= 0 || k <= 0 || i > len(vertices) || j > len(vertices) || k > len(vertices) {
					continue
				}

				triangles = append(triangles, Triangle{
					v1: vertices[i-1],
					v2: vertices[j-1],
					v3: vertices[k-1],
				})
			}
		}
	}

	return vertices, triangles, nil
}

func IsTriangleInsideCube(triangle Triangle, cube Cube) bool {
	// https://fileadmin.cs.lth.se/cs/Personal/Tomas_Akenine-Moller/code/tribox_tam.pdf
	triangleMin := Vec3{
		math.Min(triangle.v1.x, math.Min(triangle.v2.x, triangle.v3.x)),
		math.Min(triangle.v1.y, math.Min(triangle.v2.y, triangle.v3.y)),
		math.Min(triangle.v1.z, math.Min(triangle.v2.z, triangle.v3.z)),
	}
	triangleMax := Vec3{
		math.Max(triangle.v1.x, math.Max(triangle.v2.x, triangle.v3.x)),
		math.Max(triangle.v1.y, math.Max(triangle.v2.y, triangle.v3.y)),
		math.Max(triangle.v1.z, math.Max(triangle.v2.z, triangle.v3.z)),
	}

	if triangleMin.x > cube.Max.x || triangleMax.x < cube.Min.x {
		return false
	}
	if triangleMin.y > cube.Max.y || triangleMax.y < cube.Min.y {
		return false
	}
	if triangleMin.z > cube.Max.z || triangleMax.z < cube.Min.z {
		return false
	}

	return true
}

func splitCube8(cube Cube) [8]Cube {
	mid := Vec3{
		(cube.Min.x + cube.Max.x) / 2,
		(cube.Min.y + cube.Max.y) / 2,
		(cube.Min.z + cube.Max.z) / 2,
	}
	return [8]Cube{
		{Min: cube.Min, Max: mid},
		{Min: Vec3{mid.x, cube.Min.y, cube.Min.z}, Max: Vec3{cube.Max.x, mid.y, mid.z}},
		{Min: Vec3{cube.Min.x, mid.y, cube.Min.z}, Max: Vec3{mid.x, cube.Max.y, mid.z}},
		{Min: Vec3{mid.x, mid.y, cube.Min.z}, Max: Vec3{cube.Max.x, cube.Max.y, mid.z}},
		{Min: Vec3{cube.Min.x, cube.Min.y, mid.z}, Max: Vec3{mid.x, mid.y, cube.Max.z}},
		{Min: Vec3{mid.x, cube.Min.y, mid.z}, Max: Vec3{cube.Max.x, mid.y, cube.Max.z}},
		{Min: Vec3{cube.Min.x, mid.y, mid.z}, Max: Vec3{mid.x, cube.Max.y, cube.Max.z}},
		{Min: mid, Max: cube.Max},
	}
}

func calculateCenter(cube Cube) Vec3 {
	return Vec3{
		(cube.Min.x + cube.Max.x) / 2,
		(cube.Min.y + cube.Max.y) / 2,
		(cube.Min.z + cube.Max.z) / 2,
	}
}

func BuildOctree(triangles []Triangle, cube Cube, depth int, maxDepth int) *Node {
	if depth >= maxDepth {
		// Mark this cube as occupied
		return &Node{IsVoxel: true, Center: calculateCenter(cube)}
	}
	trianglesInCube := []Triangle{}
	for _, triangle := range triangles {
		if IsTriangleInsideCube(triangle, cube) {
			trianglesInCube = append(trianglesInCube, triangle)
		}
	}
	if len(trianglesInCube) == 0 {
		// Mark this cube as empty
		return nil
	}

	subCubes := splitCube8(cube)
	node := &Node{}

	// Gunakan concurrency hanya di level atas (depth < 3)
	// untuk menghindari overhead goroutine yang berlebihan
	if depth < 3 {
		var wg sync.WaitGroup
		for i, subCube := range subCubes {
			wg.Add(1)
			go func(idx int, sc Cube) {
				defer wg.Done()
				node.Children[idx] = BuildOctree(trianglesInCube, sc, depth+1, maxDepth)
			}(i, subCube)
		}
		wg.Wait()
	} else {
		// Di level dalam, jalankan secara sekuensial
		for i, subCube := range subCubes {
			node.Children[i] = BuildOctree(trianglesInCube, subCube, depth+1, maxDepth)
		}
	}

	return node
}

func collectVoxelCenters(node *Node, voxels *[]Vec3) {
	if node == nil {
		return
	}

	if node.IsVoxel {
		*voxels = append(*voxels, node.Center)
		return
	}

	for _, child := range node.Children {
		collectVoxelCenters(child, voxels)
	}
}

func exportVoxelsToOBJ(path string, voxels []Vec3, voxelSize float64) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	vertexOffset := 0
	mid := voxelSize / 2

	for _, voxel := range voxels {
		v := [8]Vec3{
			{x: voxel.x - mid, y: voxel.y - mid, z: voxel.z - mid},
			{x: voxel.x + mid, y: voxel.y - mid, z: voxel.z - mid},
			{x: voxel.x + mid, y: voxel.y + mid, z: voxel.z - mid},
			{x: voxel.x - mid, y: voxel.y + mid, z: voxel.z - mid},
			{x: voxel.x - mid, y: voxel.y - mid, z: voxel.z + mid},
			{x: voxel.x + mid, y: voxel.y - mid, z: voxel.z + mid},
			{x: voxel.x + mid, y: voxel.y + mid, z: voxel.z + mid},
			{x: voxel.x - mid, y: voxel.y + mid, z: voxel.z + mid},
		}

		for _, vertex := range v {
			_, err = fmt.Fprintf(writer, "v %f %f %f\n", vertex.x, vertex.y, vertex.z)
			if err != nil {
				return err
			}
		}

		a := vertexOffset + 1
		b := vertexOffset + 2
		c := vertexOffset + 3
		d := vertexOffset + 4
		e := vertexOffset + 5
		f := vertexOffset + 6
		g := vertexOffset + 7
		h := vertexOffset + 8

		_, err = fmt.Fprintf(writer, "f %d %d %d %d\n", a, b, c, d)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(writer, "f %d %d %d %d\n", e, f, g, h)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(writer, "f %d %d %d %d\n", a, b, f, e)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(writer, "f %d %d %d %d\n", b, c, g, f)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(writer, "f %d %d %d %d\n", c, d, h, g)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(writer, "f %d %d %d %d\n", d, a, e, h)
		if err != nil {
			return err
		}

		vertexOffset += 8
	}

	return nil
}

func voxelization(path string) {
	startTime := time.Now()

	vertices, triangles, err := readOBJ(path)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	minVec, maxVec := getBoundingBox(vertices)
	length := math.Max(maxVec.x-minVec.x, math.Max(maxVec.y-minVec.y, maxVec.z-minVec.z))
	depth := 7
	voxelSize := length / math.Pow(2, float64(depth))

	res := BuildOctree(triangles, Cube{Min: minVec, Max: maxVec}, 0, depth)
	voxels := []Vec3{}
	collectVoxelCenters(res, &voxels)

	outputPath := "voxelized.obj"
	err = exportVoxelsToOBJ(outputPath, voxels, voxelSize)
	if err != nil {
		fmt.Println("Error exporting OBJ:", err)
		return
	}

	elapsed := time.Since(startTime)
	// gunakan fungsi untuk menghasilkan output cli
	printStats(len(voxels), depth, outputPath, elapsed, res)
}

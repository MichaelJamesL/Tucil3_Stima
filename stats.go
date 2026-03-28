package main

import (
	"fmt"
	"path/filepath"
	"time"
)

type OctreeStats struct {
	NodesPerDepth   map[int]int
	PrunedPerDepth  map[int]int
}

func countOctreeNodes(node *Node, depth int, stats *OctreeStats) {
	if node == nil {
		return
	}

	stats.NodesPerDepth[depth]++

	if node.IsVoxel {
		return
	}

	for _, child := range node.Children {
		if child == nil {
			stats.PrunedPerDepth[depth+1]++
		} else {
			countOctreeNodes(child, depth+1, stats)
		}
	}
}

// printStats menampilkan seluruh statistik voxelisasi ke CLI
func printStats(voxelCount int, maxDepth int, outputPath string, elapsed time.Duration, root *Node) {
	vertexCount := voxelCount * 8  // 8 vertex per voxel (kubus)
	faceCount := voxelCount * 6    // 6 face per voxel (kubus)

	// Hitung statistik octree
	stats := &OctreeStats{
		NodesPerDepth:  make(map[int]int),
		PrunedPerDepth: make(map[int]int),
	}
	countOctreeNodes(root, 0, stats)

	absPath, _ := filepath.Abs(outputPath)

	fmt.Println("HASIL VOXELISASI 3D\n")
	fmt.Printf("Banyak voxel yang terbentuk  : %d\n", voxelCount)
	fmt.Printf("Banyak vertex yang terbentuk : %d\n", vertexCount)
	fmt.Printf("Banyak faces yang terbentuk  : %d\n", faceCount)
	fmt.Printf("Kedalaman octree             : %d\n", maxDepth)
	fmt.Printf("Lama waktu program berjalan  : %v\n", elapsed)
	fmt.Printf("Path file output             : %s\n", absPath)

	fmt.Println("\nStatistik node octree yang terbentuk:")
	for d := 0; d <= maxDepth; d++ {
		if count, ok := stats.NodesPerDepth[d]; ok {
			fmt.Printf("  %d : %d\n", d, count)
		}
	}

	fmt.Println("\nStatistik node yang tidak perlu ditelusuri (pruned):")
	for d := 1; d <= maxDepth; d++ {
		if count, ok := stats.PrunedPerDepth[d]; ok {
			fmt.Printf("  %d : %d\n", d, count)
		}
	}
}

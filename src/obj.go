package mango

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func ParseOBJ(filepath string) []*Triangle {
	return ParseOBJWithMaterial(filepath, Metal{
		Albedo:            NewSolidColourTextureAlbedo(RGB{R: 1, G: 1, B: 1}),
		Absorption:        RGB{R: 3.42, G: 2.45, B: 1.91},
		IndexOfRefraction: RGB{R: 0.18, G: 0.43, B: 1.38},
		Roughness:         0.1,
	})
}

func ParseOBJWithMaterial(filepath string, material Material) []*Triangle {
	return ParseOBJWithMaterialTransform(filepath, material, nil)
}

func ParseOBJWithMaterialTransform(filepath string, material Material, transform func(Vector3) Vector3) []*Triangle {
	file, err := os.Open(filepath)
	if err != nil {
		return nil
	}

	defer file.Close()
	if transform == nil {
		transform = func(v Vector3) Vector3 { return v }
	}

	vertices := make([]Vector3, 0)
	triangles := make([]*Triangle, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Fields(line)
		if len(split) == 0 {
			continue
		}
		t := split[0]
		if t == "v" && len(split) >= 4 {
			x, _ := strconv.ParseFloat(split[1], 64)
			y, _ := strconv.ParseFloat(split[2], 64)
			z, _ := strconv.ParseFloat(split[3], 64)

			vertices = append(vertices, transform(Vector3{x, y, z}))
			continue
		}
		if t == "f" && len(split) >= 4 {
			indices := make([]int, 0, len(split)-1)
			for _, token := range split[1:] {
				index, ok := parseOBJVertexIndex(token, len(vertices))
				if ok {
					indices = append(indices, index)
				}
			}
			for i := 1; i+1 < len(indices); i++ {
				triangles = append(triangles, NewTriangle(
					vertices[indices[0]],
					vertices[indices[i]],
					vertices[indices[i+1]],
					material,
				))
			}
			continue
		}
	}

	return triangles
}

func parseOBJVertexIndex(token string, vertexCount int) (int, bool) {
	vertexToken := strings.Split(token, "/")[0]
	index, err := strconv.Atoi(vertexToken)
	if err != nil || index == 0 {
		return 0, false
	}

	if index < 0 {
		index = vertexCount + index
	} else {
		index--
	}

	return index, index >= 0 && index < vertexCount
}

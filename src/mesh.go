package mango

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func ParseOBJ(filepath string) []*Triangle {
	file, err := os.Open(filepath)
	if err != nil {
		return nil
	}

	defer file.Close()

	vertices := make([]Vector3, 0)
	triangles := make([]*Triangle, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Fields(line)
		t := split[0]
		if t == "v" {
			x, _ := strconv.ParseFloat(split[1], 64)
			y, _ := strconv.ParseFloat(split[2], 64)
			z, _ := strconv.ParseFloat(split[3], 64)

			vertices = append(vertices, Vector3{x, y, z})
			continue
		}
		if t == "f" {
			i1, _ := strconv.ParseInt(split[1], 10, 0)
			i2, _ := strconv.ParseInt(split[2], 10, 0)
			i3, _ := strconv.ParseInt(split[3], 10, 0)

			triangles = append(triangles,
				NewTriangle(
					vertices[i1-1], vertices[i2-1], vertices[i3-1],
					Metal{
						Albedo:            NewSolidColourTextureAlbedo(RGB{R: 1, G: 1, B: 1}),
						Absorption:        RGB{R: 3.42, G: 2.45, B: 1.91},
						IndexOfRefraction: RGB{R: 0.18, G: 0.43, B: 1.38},
						Roughness:         0.1,
					},
				),
			)
			continue
		}
	}

	return triangles
}

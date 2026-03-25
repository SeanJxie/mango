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
					Lambertian{Albedo: NewSolidColourTextureAlbedo(RGB{0.5, 0.5, 0.5})},
				),
			)
			continue
		}
	}

	return triangles
}

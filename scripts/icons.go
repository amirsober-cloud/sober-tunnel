// Generates the raster sizes of sober-tunnel's code-defined tunnel mark.
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

func main() {
	dir := "engines/backpack/internal/webui/assets/icons"
	for name, size := range map[string]int{"icon-192.png": 192, "icon-512.png": 512, "icon-maskable-512.png": 512, "apple-touch-icon.png": 180} {
		img := image.NewRGBA(image.Rect(0, 0, size, size))
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				px, py := float64(x)*128/float64(size), float64(y)*128/float64(size)
				ink := false
				for _, r := range []float64{40, 20} {
					if py <= 60 && math.Abs(math.Hypot(px-64, py-60)-r) <= 4 {
						ink = true
					}
					if py >= 60 && py <= 104 && (math.Abs(px-(64-r)) <= 4 || math.Abs(px-(64+r)) <= 4) {
						ink = true
					}
				}
				if math.Abs(py-104) <= 4 && ((px >= 16 && px <= 52) || (px >= 76 && px <= 112)) {
					ink = true
				}
				c := color.RGBA{15, 23, 42, 255}
				if ink {
					c = color.RGBA{226, 232, 240, 255}
				}
				img.SetRGBA(x, y, c)
			}
		}
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			panic(err)
		}
		if err := png.Encode(f, img); err != nil {
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
	}
}

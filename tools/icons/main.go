// Command icons draws the Cais icon (a "C" made of four line-coloured arcs,
// a station at each end, on a dark tile; the same as frontend/public/icon.svg)
// as PNGs: the PWA and iPhone icons, and optionally a macOS .iconset.
//
//	go run ./tools/icons -pwa frontend/public -iconset build/Cais.iconset
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

var (
	tile  = color.NRGBA{0x17, 0x18, 0x1a, 0xff}
	white = color.NRGBA{0xff, 0xff, 0xff, 0xff}
	arcs  = []color.NRGBA{
		{0x2f, 0x7f, 0xd6, 0xff},
		{0xff, 0xc2, 0x0e, 0xff},
		{0x14, 0xa0, 0x6f, 0xff},
		{0xef, 0x33, 0x40, 0xff},
	}
)

// style places the artwork inside the canvas, in fractions of its size.
type style struct {
	inset  float64 // empty margin around the tile (macOS icons float in their canvas)
	radius float64 // tile corner radius, relative to the tile
	scale  float64 // size of the drawing relative to the tile (maskable icons keep a safe zone)
}

var (
	rounded  = style{inset: 0, radius: 14.0 / 64, scale: 1}
	square   = style{inset: 0, radius: 0, scale: 1}
	maskable = style{inset: 0, radius: 0, scale: 0.8}
	macOS    = style{inset: 100.0 / 1024, radius: 0.2237, scale: 1}
)

func main() {
	pwa := flag.String("pwa", "", "folder for the PWA and iPhone icons")
	iconset := flag.String("iconset", "", "folder for a macOS .iconset")
	flag.Parse()

	if *pwa != "" {
		write(filepath.Join(*pwa, "icon-192.png"), draw(192, rounded))
		write(filepath.Join(*pwa, "icon-512.png"), draw(512, rounded))
		write(filepath.Join(*pwa, "icon-maskable-512.png"), draw(512, maskable))
		write(filepath.Join(*pwa, "apple-touch-icon.png"), draw(180, square))
	}
	if *iconset != "" {
		for _, size := range []int{16, 32, 128, 256, 512} {
			write(filepath.Join(*iconset, fmt.Sprintf("icon_%dx%d.png", size, size)), draw(size, macOS))
			write(filepath.Join(*iconset, fmt.Sprintf("icon_%dx%d@2x.png", size, size)), draw(size*2, macOS))
		}
	}
}

// The drawing, in the 64-unit grid of icon.svg.
const (
	arcR     = 15.5
	arcW     = 7.5
	gapDeg   = 7.0
	nodeR    = 4.4
	nodeRing = 1.2 // the dark ring around each station, half of icon.svg's stroke
	startDeg = 315.0
	endDeg   = 45.0
	segDeg   = (startDeg - endDeg) / 4
)

func draw(size int, s style) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	n := float64(size)
	t0, t1 := s.inset*n, n-s.inset*n
	side := t1 - t0
	r := s.radius * side
	unit := side / 64 * s.scale
	c := t0 + side/2

	// The stations at the two ends of the "C".
	node := func(deg float64) (float64, float64) {
		a := deg * math.Pi / 180
		return c + arcR*unit*math.Cos(a), c + arcR*unit*math.Sin(a)
	}
	n1x, n1y := node(startDeg)
	n2x, n2y := node(endDeg)

	colourAt := func(x, y float64) color.NRGBA {
		for _, p := range [][2]float64{{n1x, n1y}, {n2x, n2y}} {
			d := math.Hypot(x-p[0], y-p[1])
			if d <= nodeR*unit {
				return white
			}
			if d <= (nodeR+nodeRing)*unit {
				return tile
			}
		}
		dist := math.Hypot(x-c, y-c)
		if math.Abs(dist-arcR*unit) > arcW*unit/2 {
			return tile
		}
		deg := math.Atan2(y-c, x-c) * 180 / math.Pi
		if deg < 0 {
			deg += 360
		}
		if deg > startDeg || deg < endDeg {
			return tile // the open side of the "C"
		}
		// Segment 0 starts at the top end; small gaps separate the colours.
		from := startDeg - deg
		i := int(from / segDeg)
		if i > 3 {
			i = 3
		}
		within := from - float64(i)*segDeg
		if (i > 0 && within < gapDeg/2) || (i < 3 && within > segDeg-gapDeg/2) {
			return tile
		}
		return arcs[i]
	}

	const ss = 4 // 4x4 samples per pixel for smooth edges
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			var acc [4]float64
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					x := float64(px) + (float64(sx)+0.5)/ss
					y := float64(py) + (float64(sy)+0.5)/ss
					if !inRoundRect(x, y, t0, t0, t1, t1, r) {
						continue
					}
					col := colourAt(x, y)
					acc[0] += float64(col.R)
					acc[1] += float64(col.G)
					acc[2] += float64(col.B)
					acc[3]++
				}
			}
			if acc[3] == 0 {
				continue
			}
			img.SetNRGBA(px, py, color.NRGBA{
				uint8(acc[0] / acc[3]), uint8(acc[1] / acc[3]), uint8(acc[2] / acc[3]),
				uint8(math.Round(255 * acc[3] / ss / ss)),
			})
		}
	}
	return img
}

func inRoundRect(x, y, x0, y0, x1, y1, r float64) bool {
	if x < x0 || x > x1 || y < y0 || y > y1 {
		return false
	}
	dx := math.Max(math.Max(x0+r-x, x-(x1-r)), 0)
	dy := math.Max(math.Max(y0+r-y, y-(y1-r)), 0)
	return dx*dx+dy*dy <= r*r
}

func write(path string, img image.Image) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

package desktop

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

// trayIconSize is the edge in pixels of the icon handed to the tray host. The host scales it to the
// panel; the master is too large to send, at four bytes a pixel.
const trayIconSize = 64

// pixmap is one icon image as the StatusNotifierItem specification carries it: a width, a height
// and ARGB32 pixels in network byte order, row by row.
type pixmap struct {
	Width  int32
	Height int32
	Data   []byte
}

// pixmapOf decodes a PNG and averages it down to size pixels square.
func pixmapOf(encoded []byte, size int) (pixmap, error) {
	source, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		return pixmap{}, fmt.Errorf("reading the tray icon: %w", err)
	}
	bounds := source.Bounds()
	data := make([]byte, 0, size*size*argbBytes)
	for row := range size {
		top, bottom := span(bounds.Min.Y, bounds.Dy(), row, size)
		for column := range size {
			left, right := span(bounds.Min.X, bounds.Dx(), column, size)
			data = append(data, argb(average(source, left, top, right, bottom))...)
		}
	}
	return pixmap{Width: int32(size), Height: int32(size), Data: data}, nil
}

// argbBytes is the size of one pixel as the specification sends it: alpha, red, green and blue.
const argbBytes = 4

// span answers the source pixels, from start over length, that target pixel index of count covers:
// never empty, so a target pixel always has a source.
func span(start, length, index, count int) (from, to int) {
	from = start + index*length/count
	to = start + (index+1)*length/count
	if to <= from {
		to = from + 1
	}
	return from, to
}

// average answers the mean colour of the rectangle, without premultiplied alpha.
func average(source image.Image, left, top, right, bottom int) color.NRGBA {
	var red, green, blue, alpha, count uint64
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			pixel := color.NRGBAModel.Convert(source.At(x, y)).(color.NRGBA)
			red += uint64(pixel.R)
			green += uint64(pixel.G)
			blue += uint64(pixel.B)
			alpha += uint64(pixel.A)
			count++
		}
	}
	return color.NRGBA{R: uint8(red / count), G: uint8(green / count), B: uint8(blue / count), A: uint8(alpha / count)}
}

// argb answers a pixel's bytes in the specification's order.
func argb(pixel color.NRGBA) []byte { return []byte{pixel.A, pixel.R, pixel.G, pixel.B} }

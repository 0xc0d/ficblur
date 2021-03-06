package ficblur

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/bits"
)

// Gaussian produces a blurred version of the image using a Gaussian function.
// Sigma parameter must be positive and indicates how much the image will be blurred.
// Passes parameter must be positive and indicates how many blur passes should be done.
func Gaussian(src image.Image, dst *image.RGBA, sigma, boxes int) {
	rgbaSrc, ok := src.(*image.RGBA)
	if !ok {
		rgbaSrc = new(image.RGBA)
		clone(src, rgbaSrc)
	}

	if dst == nil {
		panic("destination image must not be nil")
	}
	clone(rgbaSrc, dst)

	for _, box := range sigma2BoxDimension(sigma, boxes) {
		boxBlurHorizontal(dst, rgbaSrc, (box-1)/2)
		boxBlurVertical(rgbaSrc, dst, (box-1)/2)
	}
}

func boxBlurHorizontal(src, dst *image.RGBA, boxRadius int) {
	iarr := 1 / float64(boxRadius+boxRadius+1)
	height := src.Bounds().Max.Y - src.Bounds().Min.Y

	parallel(0, height, func(yc <-chan int) {
		for y := range yc {
			ti := src.Bounds().Min.X
			li := ti
			ri := ti + boxRadius

			fvpos := src.PixOffset(ti, y)
			lvpos := src.PixOffset(src.Bounds().Max.X-1, y)

			fvr := int(src.Pix[fvpos+0])
			fvg := int(src.Pix[fvpos+1])
			fvb := int(src.Pix[fvpos+2])
			fva := int(src.Pix[fvpos+3])

			valR := fvr * (boxRadius + 1)
			valG := fvg * (boxRadius + 1)
			valB := fvb * (boxRadius + 1)
			valA := fva * (boxRadius + 1)

			for j := 0; j < boxRadius; j++ {
				pos := src.PixOffset(ti+j, y)
				valR += int(src.Pix[pos+0])
				valG += int(src.Pix[pos+1])
				valB += int(src.Pix[pos+2])
				valA += int(src.Pix[pos+3])
			}

			for j := 0; j <= boxRadius; j++ {
				pos := src.PixOffset(ri, y)
				ri++

				valR += int(src.Pix[pos+0]) - fvr
				valG += int(src.Pix[pos+1]) - fvg
				valB += int(src.Pix[pos+2]) - fvb
				valA += int(src.Pix[pos+3]) - fva

				dst.SetRGBA(ti, y, color.RGBA{
					R: uint8(math.Round(float64(valR) * iarr)),
					G: uint8(math.Round(float64(valG) * iarr)),
					B: uint8(math.Round(float64(valB) * iarr)),
					A: uint8(math.Round(float64(valA) * iarr)),
				})

				ti++
			}

			for j := boxRadius + 1; j < src.Bounds().Max.X-boxRadius; j++ {
				ripos := src.PixOffset(ri, y)
				ri++

				lipos := src.PixOffset(li, y)
				li++

				valR += int(src.Pix[ripos+0]) - int(src.Pix[lipos+0])
				valG += int(src.Pix[ripos+1]) - int(src.Pix[lipos+1])
				valB += int(src.Pix[ripos+2]) - int(src.Pix[lipos+2])
				valA += int(src.Pix[ripos+3]) - int(src.Pix[lipos+3])

				dst.SetRGBA(ti, y, color.RGBA{
					R: uint8(math.Round(float64(valR) * iarr)),
					G: uint8(math.Round(float64(valG) * iarr)),
					B: uint8(math.Round(float64(valB) * iarr)),
					A: uint8(math.Round(float64(valA) * iarr)),
				})
				ti++
			}

			for j := src.Bounds().Max.X - boxRadius; j < src.Bounds().Max.X; j++ {
				pos := src.PixOffset(li, y)
				li++

				valR += int(src.Pix[lvpos+0]) - int(src.Pix[pos+0])
				valG += int(src.Pix[lvpos+1]) - int(src.Pix[pos+1])
				valB += int(src.Pix[lvpos+2]) - int(src.Pix[pos+2])
				valA += int(src.Pix[lvpos+3]) - int(src.Pix[pos+3])

				dst.SetRGBA(ti, y, color.RGBA{
					R: uint8(math.Round(float64(valR) * iarr)),
					G: uint8(math.Round(float64(valG) * iarr)),
					B: uint8(math.Round(float64(valB) * iarr)),
					A: uint8(math.Round(float64(valA) * iarr)),
				})
				ti++
			}
		}
	})
}

func boxBlurVertical(src, dst *image.RGBA, boxRadius int) {
	iarr := 1 / float64(boxRadius+boxRadius+1)
	width := src.Bounds().Max.X - src.Bounds().Min.X

	parallel(0, width, func(xc <-chan int) {
		for x := range xc {
			ti := src.Bounds().Min.Y
			li := ti
			ri := ti + boxRadius

			fvpos := src.PixOffset(x, ti)
			lvpos := src.PixOffset(x, src.Bounds().Max.Y-1)

			fvr := int(src.Pix[fvpos+0])
			fvg := int(src.Pix[fvpos+1])
			fvb := int(src.Pix[fvpos+2])
			fva := int(src.Pix[fvpos+3])

			valR := fvr * (boxRadius + 1)
			valG := fvg * (boxRadius + 1)
			valB := fvb * (boxRadius + 1)
			valA := fva * (boxRadius + 1)

			for j := 0; j < boxRadius; j++ {
				pos := src.PixOffset(x, ti+j)
				valR += int(src.Pix[pos+0])
				valG += int(src.Pix[pos+1])
				valB += int(src.Pix[pos+2])
				valA += int(src.Pix[pos+3])
			}

			for j := 0; j <= boxRadius; j++ {
				pos := src.PixOffset(x, ri)
				ri++

				valR += int(src.Pix[pos+0]) - fvr
				valG += int(src.Pix[pos+1]) - fvg
				valB += int(src.Pix[pos+2]) - fvb
				valA += int(src.Pix[pos+3]) - fva

				dst.SetRGBA(x, ti, color.RGBA{
					R: uint8(math.Round(float64(valR) * iarr)),
					G: uint8(math.Round(float64(valG) * iarr)),
					B: uint8(math.Round(float64(valB) * iarr)),
					A: uint8(math.Round(float64(valA) * iarr)),
				})
				ti++
			}

			for j := boxRadius + 1; j < src.Bounds().Max.Y-boxRadius; j++ {
				ripos := src.PixOffset(x, ri)
				ri++

				lipos := src.PixOffset(x, li)
				li++

				valR += int(src.Pix[ripos+0]) - int(src.Pix[lipos+0])
				valG += int(src.Pix[ripos+1]) - int(src.Pix[lipos+1])
				valB += int(src.Pix[ripos+2]) - int(src.Pix[lipos+2])
				valA += int(src.Pix[ripos+3]) - int(src.Pix[lipos+3])

				dst.SetRGBA(x, ti, color.RGBA{
					R: uint8(math.Round(float64(valR) * iarr)),
					G: uint8(math.Round(float64(valG) * iarr)),
					B: uint8(math.Round(float64(valB) * iarr)),
					A: uint8(math.Round(float64(valA) * iarr)),
				})
				ti++
			}

			for j := src.Bounds().Max.Y - boxRadius; j < src.Bounds().Max.Y; j++ {
				pos := src.PixOffset(x, li)
				li++

				valR += int(src.Pix[lvpos+0]) - int(src.Pix[pos+0])
				valG += int(src.Pix[lvpos+1]) - int(src.Pix[pos+1])
				valB += int(src.Pix[lvpos+2]) - int(src.Pix[pos+2])
				valA += int(src.Pix[lvpos+3]) - int(src.Pix[pos+3])

				dst.SetRGBA(x, ti, color.RGBA{
					R: uint8(math.Round(float64(valR) * iarr)),
					G: uint8(math.Round(float64(valG) * iarr)),
					B: uint8(math.Round(float64(valB) * iarr)),
					A: uint8(math.Round(float64(valA) * iarr)),
				})
				ti++
			}
		}
	})
}

// sigma2BoxDimension converts the standard deviation of gaussian blur (sigma)
// into dimensions of boxes for box blur.
func sigma2BoxDimension(sigma, boxes int) []int {
	wIdeal := math.Sqrt(float64(12*sigma*sigma/boxes + 1))
	wl := int(math.Floor(wIdeal))
	if wl%2 == 0 {
		wl--
	}
	wu := wl + 2

	mIdeal := (12*sigma*sigma-boxes*wl*wl+4*boxes*wl+3*boxes)/-4*wl - 4
	m := math.Round(float64(mIdeal))

	sizes := make([]int, boxes)
	for i := 0; i < boxes; i++ {
		if float64(i) < m {
			sizes[i] = wl
			continue
		}
		sizes[i] = wu
	}

	return sizes
}

// clone clones src to dst
func clone(src image.Image, dst *image.RGBA) {
	b := src.Bounds()
	dst.Pix = make([]uint8, mul3NonNeg(4, b.Dx(), b.Dy()))
	dst.Stride = 4 * b.Dx()
	dst.Rect = b

	draw.Draw(dst, b, src, b.Min, draw.Src)
}

// mul3NonNeg returns (x * y * z), unless at least one argument is negative or
// if the computation overflows the int type, in which case it returns -1.
func mul3NonNeg(x int, y int, z int) int {
	if (x < 0) || (y < 0) || (z < 0) {
		return -1
	}
	hi, lo := bits.Mul64(uint64(x), uint64(y))
	if hi != 0 {
		return -1
	}
	hi, lo = bits.Mul64(lo, uint64(z))
	if hi != 0 {
		return -1
	}
	a := int(lo)
	if (a < 0) || (uint64(a) != lo) {
		return -1
	}
	return a
}

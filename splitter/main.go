package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/lucasb-eyer/go-colorful"
)

func main() {
	files, err := filepath.Glob("*.png")

	if err != nil {
		panic(err)
	}

	crop := func(src image.Image, rect image.Rectangle) image.Image {
		dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
		draw.Draw(dst, dst.Bounds(), src, rect.Min, draw.Src)

		return dst
	}
	white, _ := colorful.MakeColor(color.White)

	exec := func(writer func(img image.Image, cut, center bool)) {
		for i := range files {
			func() {
				fp, err := os.Open(files[i])

				if err != nil {
					panic(err)
				}
				defer fp.Close()

				img, err := png.Decode(fp)

				if err != nil {
					panic(err)
				}

				width, height := img.Bounds().Dx(), img.Bounds().Dy()

				checkBand := func(mid int) int {
					cnt := 0
					for j := mid - 2; j <= mid+2; j++ {
						for i := 0; i < height; i++ {
							col, _ := colorful.MakeColor(img.At(j, i))

							if col.DistanceRgb(white) > 0.2 {
								cnt++
							}
						}
					}
					return cnt
				}

				if checkBand(width/2) < 20 {
					right := crop(img, image.Rect(width/2, 0, width, height))

					writer(right, true, false)

					left := crop(img, image.Rect(0, 0, width/2, height))

					writer(left, true, false)

					return
				}

				if checkBand(width/4)+checkBand(width*3/4) < 20 {
					left := crop(img, image.Rect(width/4, 0, width*3/4, height))
					writer(left, true, true)

					return
				}

				writer(img, false, true)
			}()
		}
	}

	isWhitePlain := func(img image.Image) bool {
		width, height := img.Bounds().Dx(), img.Bounds().Dy()

		cnt := 0
		for j := 0; j < width; j++ {
			for i := 0; i < height; i++ {
				col, _ := colorful.MakeColor(img.At(j, i))

				if col.DistanceRgb(white) > 0.2 {
					cnt++

					if cnt > 20 {
						return false
					}
				}
			}

		}

		return true
	}

	calcLeftSpace := func(img image.Image) int {
		width, height := img.Bounds().Dx(), img.Bounds().Dy()

		ln := 0
		for j := 0; j < width; j++ {
			cnt := 0
			for i := 0; i < height; i++ {
				col, _ := colorful.MakeColor(img.At(j, i))

				if col.DistanceRgb(white) > 0.2 {
					cnt++
				}
			}

			if cnt > 20 {
				break
			}

			ln++
		}

		return ln
	}

	calcRightSpace := func(img image.Image) int {
		width, height := img.Bounds().Dx(), img.Bounds().Dy()

		ln := 0
		for j := width - 1; j >= 0; j-- {
			cnt := 0
			for i := 0; i < height; i++ {
				col, _ := colorful.MakeColor(img.At(j, i))

				if col.DistanceRgb(white) > 0.2 {
					cnt++
				}
			}

			if cnt > 20 {
				break
			}

			ln++
		}

		return ln
	}

	ux, lx := int(1e9), int(1e9) //
	checker := func(img image.Image, cut, center bool) {
		if !cut {
			return
		}

		left, right := calcLeftSpace(img), calcRightSpace(img)

		if right > left {
			left, right = right, left
		}
		// right < left

		if right < lx {
			lx = right
		}
		if left < ux {
			ux = left
		}

	}
	exec(checker)

	mx := ux - lx

	if err := os.MkdirAll("./out", 0770); err != nil {
		panic(err)
	}

	counter := 0
	writer := func(img image.Image, cut, center bool) {
		save := func(img image.Image) {
			if isWhitePlain(img) {
				return
			}

			writer, err := os.Create(fmt.Sprintf("out/%03d.png", counter))
			counter++

			if err != nil {
				panic(err)
			}
			defer writer.Close()

			if err := png.Encode(writer, img); err != nil {
				panic(err)
			}
		}

		if !cut {
			save(img)

			return
		}

		width, height := img.Bounds().Dx(), img.Bounds().Dy()
		newWidth := width - mx
		if center {
			save(crop(img, image.Rect(mx/2, 0, newWidth+mx/2, height)))

			return
		}

		ls, rs := calcLeftSpace(img), calcRightSpace(img)

		if ls < rs {
			save(crop(img, image.Rect(0, 0, newWidth, height)))
		} else {
			save(crop(img, image.Rect(mx, 0, width, height)))
		}
	}

	exec(writer)
}

package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kbinani/screenshot"
)

func diffImage(img1, img2 *image.RGBA) int {
	if img1 == nil {
		if img2 == nil {
			return 0
		}

		size := img2.Bounds().Size()

		return size.X * size.Y
	}
	if img2 == nil {
		size := img1.Bounds().Size()

		return size.X * size.Y
	}

	size1 := img1.Bounds().Size()
	size2 := img2.Bounds().Size()
	if !size1.Eq(size2) {
		return 10000000 // not correct
	}

	counter := 0
	bounds := img1.Bounds()
	for x := bounds.Min.X; x <= bounds.Max.X; x++ {
		for y := bounds.Min.Y; y <= bounds.Max.Y; y++ {
			if img1.At(x, y) != img2.At(x, y) {
				counter++
			}
		}
	}

	return counter
}

func saveImage(img *image.RGBA, path string) {
	fp, err := os.Create(path)

	if err != nil {
		fmt.Println(err)

		return
	}
	defer fp.Close()
	if err := png.Encode(fp, img); err != nil {
		fmt.Println(err)

		return
	}
}

func main() {
	rawImageCh := make(chan *image.RGBA, 500)
	filteredImageCh := make(chan *image.RGBA, 500)

	go func() {
		sig := make(chan os.Signal, 1)

		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		time.Sleep(10 * time.Second)
		fmt.Println("start")
		for {
			select {
			case <-sig:
				close(rawImageCh)
				return
			default:
			}

			img, err := screenshot.CaptureDisplay(0)

			if err != nil {
				log.Println(err)
				continue
			}

			rawImageCh <- img
		}
	}()

	threshold := 10000
	go func() {
		var prev *image.RGBA
		for {
			select {
			case img, ok := <-rawImageCh:
				if !ok {
					close(filteredImageCh)
					return
				}

				if diffImage(prev, img) > threshold {
					filteredImageCh <- img
				}
				prev = img
			}
		}
	}()

	counter := 0
	for {
		select {
		case img, ok := <-filteredImageCh:
			if !ok {
				return
			}

			saveImage(img, fmt.Sprintf("%03d.png", counter))
			counter++
			fmt.Println(counter)
		}
	}
}

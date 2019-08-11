package main

import (
	"testing"
	"time"

	"github.com/kbinani/screenshot"
)

func BenchmarkSS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := screenshot.CaptureDisplay(0)

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompare(b *testing.B) {
	b.StopTimer()
	img1, err1 := screenshot.CaptureDisplay(0)
	time.Sleep(5 * time.Second)
	img2, err2 := screenshot.CaptureDisplay(0)

	if err1 != nil || err2 != nil {
		b.Fatal(err1, err2)
	}

	if len(img1.Pix) != len(img2.Pix) {
		b.Fatal("length differs")
	}

	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		counter := 0

		bounds := img1.Bounds()
		for x := bounds.Min.X; x <= bounds.Max.X; x++ {
			for y := bounds.Min.Y; y <= bounds.Max.Y; y++ {
				if img1.At(x, y) != img2.At(x, y) {
					counter++
				}
			}
		}
		// for i := range img1.Pix {
		// 	if img1.Pix[i] != img2.Pix[i] {
		// 		same = false
		// 		break
		// 	}
		// }

		b.Log(counter)
	}
}

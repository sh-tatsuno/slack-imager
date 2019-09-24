package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/image/draw"
)

const (
	// ExitCodeOK : exit code
	ExitCodeOK = 0
	// ExitCodeError : error code
	ExitCodeError = 1
	slackX        = 128
	slackY        = 128
)

func usage() {
	io.WriteString(os.Stderr, usageText)
	flag.PrintDefaults()
}

const usageText = `this is image convert library by go.
In normal usage, you should set -d for directory and -i for input extension.
You also have to set output extension by -o.
You can also set maximum nuber you want to convert by set n.
current available extensions are jpg, jpeg, png, and gif.
Example:
    gophoto -d dir -i .png -o .jpeg -n 10
`

func main() {
	os.Exit(run(os.Args[1:]))
}
func run(args []string) int {
	var input, output string
	// args
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.Usage = usage
	flags.StringVar(&input, "i", "", "input extension.")
	flags.StringVar(&output, "o", "", "output extension.")
	flags.Parse(args)
	if input == "" {
		usage()
		fmt.Fprintf(os.Stderr, "Expected dir path\n")
		return ExitCodeError
	}
	// open
	file, err := os.Open(input)
	defer file.Close()
	if err != nil {
		return ExitCodeError
	}
	imgSrc, _, err := image.Decode(file)
	if err != nil {
		return ExitCodeError
	}
	imgDst := image.NewRGBA(image.Rect(0, 0, slackX, slackY))
	draw.CatmullRom.Scale(imgDst, imgDst.Bounds(), imgSrc, imgSrc.Bounds(), draw.Over, nil)
	rawPath := output + ".png"
	dst, err := os.Create(rawPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitCodeError
	}
	defer dst.Close()
	err = png.Encode(dst, imgDst)
	if err != nil {
		return ExitCodeError
	}
	// gray
	grayPath := output + "-gray.png"
	dstGray, err := os.Create(grayPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitCodeError
	}
	defer dstGray.Close()
	imgGray := imageGray(imgDst)
	err = png.Encode(dstGray, imgGray)
	if err != nil {
		return ExitCodeError
	}
	// nega
	negaPath := output + "-nega.png"
	dstNega, err := os.Create(negaPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitCodeError
	}
	defer dstNega.Close()
	imgNega := imageNega(imgDst)
	err = png.Encode(dstNega, imgNega)
	if err != nil {
		return ExitCodeError
	}
	// moza
	mozaPath := output + "-moza.png"
	dstMoza, err := os.Create(mozaPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitCodeError
	}
	defer dstMoza.Close()
	imgMoza := imageMoza(imgDst)
	err = png.Encode(dstMoza, imgMoza)
	if err != nil {
		return ExitCodeError
	}
	imgConv(output+"-sepia.png", imgDst, imageSepia)
	imgConv(output+"-smog.png", imgDst, imageSmog)
	imgConv(output+"-red.png", imgDst, imageRed)
	imgConv(output+"-blue.png", imgDst, imageBlue)
	imgConv(output+"-green.png", imgDst, imageGreen)
	imgConv(output+"-sb.png", imgDst, imageSkyBlue)
	// imgConvColorCode(output+"-ffb6c1.png", imgDst, "ffb6c1")
	urls := getCodes()
	for _, u := range urls {
		imgConvColorCode(output+"-"+u+".png", imgDst, u)
	}
	return ExitCodeOK
}
func imgConv(path string, imgSrc *image.RGBA,
	f func(img *image.RGBA) *image.RGBA) error {
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()
	imgDst := f(imgSrc)
	err = png.Encode(dst, imgDst)
	if err != nil {
		return err
	}
	return nil
}
func imageGray(img *image.RGBA) *image.Gray16 {
	bounds := img.Bounds()
	dest := image.NewGray16(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.Gray16Model.Convert(img.At(x, y))
			gray, _ := c.(color.Gray16)
			dest.Set(x, y, gray)
		}
	}
	return dest
}
func imageSepia(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.RGBAModel.Convert(img.At(x, y))
			col := c.(color.RGBA)
			avg := (float32(col.R) + float32(col.G) + float32(col.B)) / 3
			cg := float32(col.G) * 0.7
			cb := float32(col.B) * 0.4
			dest.Set(x, y, color.RGBA{uint8(avg), uint8(cg), uint8(cb), col.A})
		}
	}
	return dest
}
func imageNega(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.RGBAModel.Convert(img.At(x, y))
			col := c.(color.RGBA)
			cr := uint8(255 - int(col.R))
			cg := uint8(255 - int(col.G))
			cb := uint8(255 - int(col.B))
			dest.Set(x, y, color.RGBA{cr, cg, cb, col.A})
		}
	}
	return dest
}
func imageMoza(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)
	block := 5
	for y := bounds.Min.Y + (block-1)/2; y < bounds.Max.Y; y = y + block {
		for x := bounds.Min.X + (block-1)/2; x < bounds.Max.X; x = x + block {
			var cr, cg, cb float32
			var alpha uint8
			for j := y - (block-1)/2; j <= y+(block-1)/2; j++ {
				for i := x - (block-1)/2; i <= x+(block-1)/2; i++ {
					if i >= 0 && j >= 0 && i < bounds.Max.X && j < bounds.Max.Y {
						c := color.RGBAModel.Convert(img.At(i, j))
						col := c.(color.RGBA)
						cr += float32(col.R)
						cg += float32(col.G)
						cb += float32(col.B)
						alpha = col.A
					}
				}
			}
			cr = cr / float32(block*block)
			cg = cg / float32(block*block)
			cb = cb / float32(block*block)
			for j := y - (block-1)/2; j <= y+(block-1)/2; j++ {
				for i := x - (block-1)/2; i <= x+(block-1)/2; i++ {
					if i >= 0 && j >= 0 && i < bounds.Max.X && j < bounds.Max.Y {
						dest.Set(i, j, color.RGBA{uint8(cr), uint8(cg), uint8(cb), alpha})
					}
				}
			}
		}
	}
	return dest
}
func imageSmog(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.RGBAModel.Convert(img.At(x, y))
			col := c.(color.RGBA)
			cr := float32(col.R) * 0.2
			cg := float32(col.G) * 0.2
			cb := float32(col.B) * 0.2
			dest.Set(x, y, color.RGBA{uint8(cr), uint8(cg), uint8(cb), col.A})
		}
	}
	return dest
}
func imageRed(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.RGBAModel.Convert(img.At(x, y))
			col := c.(color.RGBA)
			cr := float32(col.R)*0.5 + 100
			cg := float32(col.G) * 0.5
			cb := float32(col.B) * 0.5
			dest.Set(x, y, color.RGBA{uint8(cr), uint8(cg), uint8(cb), col.A})
		}
	}
	return dest
}
func imageBlue(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.RGBAModel.Convert(img.At(x, y))
			col := c.(color.RGBA)
			cr := float32(col.R) * 0.5
			cg := float32(col.G) * 0.5
			cb := float32(col.B)*0.5 + 100
			dest.Set(x, y, color.RGBA{uint8(cr), uint8(cg), uint8(cb), col.A})
		}
	}
	return dest
}
func imageGreen(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.RGBAModel.Convert(img.At(x, y))
			col := c.(color.RGBA)
			cr := float32(col.R) * 0.5
			cg := float32(col.G)*0.5 + 100
			cb := float32(col.B) * 0.5
			dest.Set(x, y, color.RGBA{uint8(cr), uint8(cg), uint8(cb), col.A})
		}
	}
	return dest
}
func imageSkyBlue(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.RGBAModel.Convert(img.At(x, y))
			col := c.(color.RGBA)
			cr := float32(col.R) * 0.5
			cg := float32(col.G)*0.5 + 100
			cb := float32(col.B)*0.5 + 100
			dest.Set(x, y, color.RGBA{uint8(cr), uint8(cg), uint8(cb), col.A})
		}
	}
	return dest
}
func rgb16(in string) (float32, float32, float32, error) {
	conv, err := strconv.ParseInt(in, 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	src := float32(conv)
	r := float32(conv / (256 * 256))
	src -= float32(int(r) * (256 * 256))
	g := float32(src / 256)
	b := float32(src - float32(int(g)*256))
	return r, g, b, nil
}
func imageColor(colorCode string, img *image.RGBA) (*image.RGBA, error) {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)
	r, g, b, err := rgb16(colorCode)
	if err != nil {
		return nil, err
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.RGBAModel.Convert(img.At(x, y))
			col := c.(color.RGBA)
			cr := (float32(col.R) + r) * 0.5
			cg := (float32(col.G) + g) * 0.5
			cb := (float32(col.B) + b) * 0.5
			dest.Set(x, y, color.RGBA{uint8(cr), uint8(cg), uint8(cb), col.A})
		}
	}
	return dest, nil
}
func imgConvColorCode(path string, imgSrc *image.RGBA, colorCode string) error {
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()
	imgDst, err := imageColor(colorCode, imgSrc)
	if err != nil {
		return err
	}
	err = png.Encode(dst, imgDst)
	if err != nil {
		return err
	}
	return nil
}
func getCodes() []string {
	doc, err := goquery.NewDocument("https://www.colordic.org/")
	if err != nil {
		fmt.Print("url scarapping failed")
	}
	urls := []string{}
	doc.Find("table > tbody > tr > td").Each(func(_ int, s *goquery.Selection) {
		url, _ := s.Attr("style")
		urls = append(urls, url[18:])
		fmt.Println(url[18:])
	})
	return urls
}

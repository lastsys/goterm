package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
)

// According to http://hitmen.c02.at/temp/palstuff/
var C64Colors = [16]color.RGBA{
	{0x00, 0x00, 0x00, 0xFF}, //  0 - black
	{0xFF, 0xFF, 0xFF, 0xFF}, //  1 - white
	{0x67, 0x37, 0x2B, 0xFF}, //  2 - red
	{0x6F, 0xA3, 0xB1, 0xFF}, //  3 - cyan
	{0x6F, 0x3C, 0x85, 0xFF}, //  4 - purple
	{0x58, 0x8C, 0x43, 0xFF}, //  5 - green
	{0x34, 0x28, 0x79, 0xFF}, //  6 - blue
	{0xB7, 0xC6, 0x6E, 0xFF}, //  7 - yellow
	{0x6F, 0x4F, 0x25, 0xFF}, //  8 - orange
	{0x42, 0x39, 0x00, 0xFF}, //  9 - brown
	{0x99, 0x66, 0x59, 0xFF}, // 10 - light red
	{0x43, 0x43, 0x43, 0xFF}, // 11 - dark grey
	{0x6B, 0x6B, 0x6B, 0xFF}, // 12 - grey
	{0x9A, 0xD1, 0x83, 0xFF}, // 13 - light green
	{0x6B, 0x5E, 0xB4, 0xFF}, // 14 - light blue
	{0x95, 0x95, 0x95, 0xFF}, // 15 - light grey
}

type RawFont = []byte

func readRawFont(filename string) RawFont {
	bytes, err := ioutil.ReadFile("./font/" + filename)
	if err != nil {
		panic(err)
	}
	return bytes[2:]
}

func pixelColor(row byte, pixel uint, backgroundColor color.Color, foregroundColor color.Color) color.Color {
	var value = row & (1 << pixel)
	if value > 0 {
		return foregroundColor
	}
	return backgroundColor
}

func buildImage(lowerCase RawFont, upperCase RawFont) *image.RGBA {
	const charSize = 8
	const charCount = 128
	var img = image.NewRGBA(image.Rect(0, 0, charSize*charCount*2,
		charSize*len(C64Colors)*len(C64Colors)))
	var x, y int

	for k, row := range upperCase {
		for i, backgroundColor := range C64Colors {
			for j, foregroundColor := range C64Colors {
				y = (i+j*len(C64Colors))*charSize + k%charSize
				x = (k / charSize) * charSize
				for p := 0; p < charSize; p++ {
					img.Set(x+p, y, pixelColor(row, uint(charSize-p-1),
						backgroundColor, foregroundColor))
				}
			}
		}
	}

	for k, row := range lowerCase {
		for i, backgroundColor := range C64Colors {
			for j, foregroundColor := range C64Colors {
				y = (i+j*len(C64Colors))*charSize + k%charSize
				x = (k/charSize + charCount) * charSize
				for p := 0; p < charSize; p++ {
					img.Set(x+p, y, pixelColor(row, uint(charSize-p-1),
						backgroundColor, foregroundColor))
				}
			}
		}
	}

	return img
}

func saveImage(img *image.RGBA) {
	f, err := os.Create("./client/font.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	png.Encode(w, img)
	w.Flush()
}

func GenerateFont() {
	dir, _ := os.Getwd()
	fmt.Println("Current working directory: ", dir)
	lowerCase := readRawFont("c64_lower.64c")
	upperCase := readRawFont("c64_upper.64c")
	img := buildImage(lowerCase, upperCase)
	saveImage(img)
}

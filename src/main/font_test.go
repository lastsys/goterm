package main

import (
	"fmt"
	"os"
	"testing"
)

func TestReadFont(t *testing.T) {
	dir, _ := os.Getwd()
	fmt.Println(dir)
	lowerCase := readRawFont("c64_lower.64c")
	upperCase := readRawFont("c64_upper.64c")
	img := buildImage(lowerCase, upperCase)
	saveImage(img)
}

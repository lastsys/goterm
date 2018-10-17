package main

import (
	"encoding/json"
)

const WIDTH = 40
const HEIGHT = 25

type Character struct {
	Char       string `json:"char"`
	Foreground uint   `json:"fg"`
	Background uint   `json:"bg"`
}

type Position struct {
	Row    uint `json:"row"`
	Column uint `json:"col"`
}

type Buffer struct {
	Chars          [HEIGHT][WIDTH]Character `json:"chars"`
	CursorPosition Position                 `json:"cursorPosition"`
}

func (b *Buffer) Encode() []byte {
	bytes, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (b *Buffer) PutChar(char Character, row uint, col uint) {
	b.Chars[row][col] = char
}

func (b *Buffer) PutCharAtCursor(char Character) {
	b.Chars[b.CursorPosition.Row][b.CursorPosition.Column] = char
	b.CursorPosition.Column++
	if b.CursorPosition.Column >= WIDTH {
		b.CursorPosition.Column = 0
		b.CursorPosition.Row++
	}
	if b.CursorPosition.Row >= HEIGHT {
		b.CursorPosition.Row = 0
	}
}

func (b *Buffer) Print(text string, row uint, col uint, fg uint, bg uint) {
	var y = row
	var x = col

	for _, c := range text {
		b.PutChar(Character{string(c), fg, bg}, y, x)
		x++
		if x >= WIDTH {
			x = 0
			y++
		}
		if y >= HEIGHT {
			y = 0
		}
	}
	b.CursorPosition.Row = y
	b.CursorPosition.Column = x
}

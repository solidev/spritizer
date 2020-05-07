package main

import (
	"image"
)

type Sprite struct {
	W    int
	H    int
	X    int
	Y    int
	Name string
	File string
	Data image.Image
}

type Collection struct {
	W       int
	H       int
	Fill    int
	Sprites []*Sprite
}

package main

import (
	"github.com/sirupsen/logrus"
	"math"
	"sort"
)

type BoxSpace struct {
	X int
	Y int
	W int
	H int
}

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// Organize images in a box.
// Code from https://github.com/mapbox/potpack
func (coll *Collection) Organize() error {
	logrus.Debug("Organizing images")
	// calculate total box area and maximum box width
	area := 0
	maxWidth := 0

	for _, b := range coll.Sprites {
		logrus.Debugf("Processing sprite %s of size %d x %d", b.Name, b.W, b.H)
		area += b.W * b.H
		maxWidth = Max(maxWidth, b.W)
	}

	// sort the boxes for insertion by height, descending
	less := func(i, j int) bool {
		return coll.Sprites[i].H > coll.Sprites[j].H
	}
	sort.SliceStable(coll.Sprites, less)
	// aim for a squarish resulting container,
	// slightly adjusted for sub-100% space utilization
	startWidth := int(math.Max(math.Ceil(math.Sqrt(float64(area)/0.95)), float64(maxWidth)))
	// start with a single empty space, unbounded at the bottom
	spaces := []*BoxSpace{&BoxSpace{X: 0, Y: 0, W: startWidth, H: math.MaxInt64}}
	width := 0
	height := 0

	for _, box := range coll.Sprites {
		logrus.Debugf("Placing box %s (size %d x %d)\n", box.Name, box.W, box.H)
		for i, s := range spaces {
			logrus.Debugf("Space %d : x=%d y=%d w=%d h=%d",
				i, s.X, s.Y, s.W, s.H)
		}
		for i := len(spaces) - 1; i >= 0; i-- {
			// look through spaces backwards so that we check smaller spaces first
			space := spaces[i]
			if (box.W > space.W) || (box.H > space.H) {
				continue
			}
			// found the space; add the box to its top-left corner
			// |-------|-------|
			// |  box  |       |
			// |_______|       |
			// |         space |
			// |_______________|
			box.X = space.X
			box.Y = space.Y
			height = Max(height, box.Y+box.H)
			width = Max(width, box.X+box.W)

			if (box.W == space.W) && (box.H == space.H) {
				// space matches the box exactly; remove it
				var last *BoxSpace
				last, spaces = spaces[len(spaces)-1], spaces[:len(spaces)-1]
				if i < len(spaces) {
					spaces[i] = last
				}
			} else if box.H == space.H {
				// space matches the box height; update it accordingly
				// |-------|---------------|
				// |  box  | updated space |
				// |_______|_______________|
				space.X += box.W
				space.W -= box.W
			} else if box.W == space.W {
				// space matches the box width; update it accordingly
				// |---------------|
				// |      box      |
				// |_______________|
				// | updated space |
				// |_______________|
				space.Y += box.H
				space.H -= box.H
			} else {
				// otherwise the box splits the space into two spaces
				// |-------|-----------|
				// |  box  | new space |
				// |_______|___________|
				// | updated space     |
				// |___________________|
				spaces = append(spaces, &BoxSpace{
					X: space.X + box.W,
					Y: space.Y,
					W: space.W - box.W,
					H: box.H,
				})
				space.Y += box.H
				space.H -= box.H
			}
			break
		}
	}
	coll.W = width
	coll.H = height
	if width*height != 0 {
		coll.Fill = (area * 100) / (width * height)
	} else {
		logrus.Error("Zero fill : width * height = 0")
		coll.Fill = -1
	}
	return nil
}

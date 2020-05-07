package main

import (
	"bytes"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"github.com/urfave/cli/v2"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func loadImage(filename string) (int, int, image.Image, error) {
	// Read image from file that already exists
	imageFile, err := os.Open(filename)
	if err != nil {
		return 0, 0, nil, nil
	}
	defer imageFile.Close()

	// Calling the generic image.Decode() will tell give us the data
	// and type of image it is as a string. We expect "png"
	config, _, err := image.DecodeConfig(imageFile)
	if err != nil {
		return 0, 0, nil, err
	}
	_, _ = imageFile.Seek(0, 0)
	imageData, _, err := image.Decode(imageFile)
	if err != nil {
		return config.Width, config.Height, imageData, err
	}
	logrus.Debugf("Loaded image %s with %d x %d size", filename, config.Width, config.Height)
	return config.Width, config.Height, imageData, nil
}

func loadSvgOkSvg(filename string, resize int) (int, int, image.Image, error) {
	logrus.Debugf("Exporting %s to png with oksvg", filename)
	icon, errSvg := oksvg.ReadIcon(filename, oksvg.StrictErrorMode)
	if errSvg != nil {
		logrus.Warningf("oksvg not able to process %s : %v", filename, errSvg)
		return 0, 0, nil, errSvg
	}
	var w, h int
	if resize == 0 {
		w, h = int(icon.ViewBox.W), int(icon.ViewBox.H)
	} else {
		w = resize
		h = int(icon.ViewBox.H * float64(resize) / icon.ViewBox.W)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scannerGV := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scannerGV)
	icon.SetTarget(0, 0, float64(w), float64(h))
	icon.Draw(raster, 1.0)
	return w, h, img, nil
}

func loadSvgInkscape(filename string, resize int) (int, int, image.Image, error) {
	pngout := filepath.Join(os.TempDir(), filepath.Base(filename)+".png")
	defer os.Remove(pngout)
	var cmd *exec.Cmd
	if resize == 0 {
		logrus.Debugf("Exporting %s to png with inkscape", filename)
		cmd = exec.Command("inkscape", "--export-png", pngout, "--export-background-opacity", "0", filename)
	} else {
		logrus.Debugf("Exporting %s to png with inkscape width %d", filename, resize)
		cmd = exec.Command("inkscape", "--export-png", pngout, "-w", strconv.Itoa(resize), "--export-background-opacity", "0", filename)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logrus.Warning(stdout.String())
		logrus.Warning(stderr.String())
		return 0, 0, nil, err
	}
	return loadImage(pngout)
}

func loadSvg(filename string, resize int, preferInkcape bool) (int, int, image.Image, error) {
	var w, h int
	var img image.Image
	var err error
	if !preferInkcape {
		w, h, img, err = loadSvgOkSvg(filename, resize)
		if err == nil {
			return w, h, img, err
		}
	}
	w, h, img, err = loadSvgInkscape(filename, resize)
	if err == nil {
		return w, h, img, err
	}
	return 0, 0, nil, err
}

func (coll *Collection) Load(c *cli.Context) error {
	inputdir := c.Args().Get(0)
	files, err := ioutil.ReadDir(inputdir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(inputdir, file.Name())
		ext := strings.ToLower(filepath.Ext(file.Name()))
		var w, h int
		var data image.Image
		var erri error
		resize := c.Int("resize")
		// Load image
		if ext == ".svg" {
			logrus.Infof("Importing svg %s", file.Name())
			w, h, data, erri = loadSvg(path, resize, c.Bool("inkscape"))
			if erri != nil {
				logrus.Errorf("Error importing %s : %v", path, erri)
				continue
			}
		} else if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
			logrus.Infof("Importing png/jpg/gif %s", file.Name())
			w, h, data, erri = loadImage(path)
			if erri != nil {
				logrus.Errorf("Error importing %s : %v", path, erri)
				continue
			}
			if resize != 0 {
				logrus.Debugf("Resizing image %s to %d x %d", path, resize, (resize*h)/w)
				data = imaging.Resize(data, resize, (resize*h)/w, imaging.Lanczos)
				h = (resize * h) / w
				w = resize
			}
		}
		coll.Sprites = append(coll.Sprites, &Sprite{
			W:    w,
			H:    h,
			X:    0,
			Y:    0,
			Name: strings.Replace(file.Name(), ext, "", 1),
			File: path,
			Data: data,
		})

	}

	return nil
}

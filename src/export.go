package main

import (
	"github.com/Masterminds/sprig"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"text/template"
)

const jsonTemplate = `
{{- $last := sub ( len .Sprites ) 1 -}}
{	
{{- range $i, $c := .Sprites }}
	"{{ $c.Name }}": {
		"height": {{$c.H}},
		"width": {{$c.W}},
		"pixelRatio": 1,
		"x": {{$c.X}},
		"y": {{$c.Y}}
	}{{ if ne $i $last }},{{ end }}
{{- end }}
}
`

func (coll *Collection) Export(c *cli.Context) error {
	logrus.Debug("Exporting collection")
	img := image.NewRGBA(image.Rect(0, 0, coll.W, coll.H))
	for _, s := range coll.Sprites {
		pos := image.Point{X: s.X, Y: s.Y}
		r := image.Rectangle{Min: pos, Max: pos.Add(image.Point{X: s.W, Y: s.H})}
		sr := image.Rectangle{Min: image.Point{}, Max: pos.Add(image.Point{X: s.W, Y: s.H})}
		draw.Draw(img, r, s.Data, sr.Min, draw.Src)
	}
	f, err := os.Create(filepath.Join(c.Args().Get(1), c.String("name")+".png"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	err = png.Encode(f, img)
	if err != nil {
		logrus.Error(err)
		return err
	}
	// Return style using template
	var t *template.Template
	if c.String("template") != "" {
		t, err = template.New("base").Funcs(sprig.TxtFuncMap()).ParseFiles(c.String("template"))
		if err != nil {
			logrus.Errorf("Unable to parse template %v", err)
			return err
		}
	} else {
		t, err = template.New("default").Funcs(sprig.TxtFuncMap()).Parse(jsonTemplate)
	}
	f, err = os.Create(filepath.Join(c.Args().Get(1), c.String("name")+c.String("ext")))
	if err != nil {
		logrus.Error(err)
		return err
	}
	err = t.Execute(f, coll)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

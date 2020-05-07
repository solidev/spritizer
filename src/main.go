package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {

	app := &cli.App{
		Name:  "spritizer",
		Usage: "Create sprites from a directory of images",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "debug",
				Aliases:  nil,
				Usage:    "Set logging to debug",
				Required: false,
				Value:    false,
			},
			&cli.BoolFlag{
				Name:     "quiet",
				Aliases:  nil,
				Usage:    "Only log errors",
				Required: false,
				Value:    false,
			},
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:      "gen",
				Aliases:   []string{"g"},
				Usage:     "Generate sprites",
				UsageText: "gen - generate sprites",
				ArgsUsage: "INPUT_DIR OUTPUT_DIR",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "inkscape",
						Aliases:  nil,
						Usage:    "Use inkscape if available to process svg",
						Required: false,
						Value:    false,
					},
					&cli.IntFlag{
						Name:     "resize",
						Aliases:  nil,
						Usage:    "Resize to [resize] pixels width",
						Required: false,
						Hidden:   false,
					},
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "Output name",
						Required: false,
						Value:    "sprite",
					},
					&cli.StringFlag{
						Name:     "template",
						Aliases:  []string{"t"},
						Usage:    "Go template to use for textual data",
						Required: false,
						Value:    "",
					},
					&cli.StringFlag{
						Name:     "ext",
						Aliases:  []string{"e"},
						Usage:    "Summary file extension",
						Required: false,
						Value:    ".json",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Args().Len() < 2 {
						logrus.Error("Missing arguments")
						cli.ShowAppHelp(c)
						os.Exit(1)
					}
					if c.Bool("debug") {
						logrus.SetLevel(logrus.DebugLevel)
					}
					if c.Bool("quiet") {
						logrus.SetLevel(logrus.ErrorLevel)
					}

					coll := Collection{}
					err := coll.Load(c)
					if err != nil {
						logrus.Errorf("Error loading images : %v", err)
						return err
					}
					err = coll.Organize()
					if err != nil {
						logrus.Errorf("Unable to organize images : %v", err)
						return err
					}
					err = coll.Export(c)
					if err != nil {
						logrus.Errorf("Unable to export images : %v", err)
						return err
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

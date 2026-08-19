package main

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    false,
	ReportTimestamp: true,
	TimeFormat:      time.TimeOnly,
	Level:           log.DebugLevel,
	Prefix:          "Patcher",
})

var (
	info      map[string]interface{}
	directory string
	assets    string
	ipa       string
	appName   string
	iconURL   string
	iconZip   string
	output    string
)

func main() {
	app := &cli.App{
		Name:  "retribution-patcher",
		Usage: "Patches the Discord IPA with Retribution icons and sideloading fixes.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "name",
				Usage:       "Display name for the patched app",
				Value:       "Retribution",
				Destination: &appName,
			},
			&cli.StringFlag{
				Name:        "icon-url",
				Usage:       "URL to an icon zip archive",
				Value:       "",
				Destination: &iconURL,
			},
			&cli.StringFlag{
				Name:        "icon-zip",
				Usage:       "Local path to an icon zip archive",
				Value:       "",
				Destination: &iconZip,
			},
			&cli.StringFlag{
				Name:        "output",
				Usage:       "Output IPA file name",
				Value:       "Retribution.ipa",
				Destination: &output,
			},
		},
		Action: func(context *cli.Context) error {
			ipa = context.Args().Get(0)

			if ipa == "" {
				logger.Error("Please provide a path to the IPA.")
				os.Exit(1)
			}

			if iconZip == "" && iconURL == "" {
				logger.Error("Please provide an icon source with --icon-zip or --icon-url.")
				os.Exit(1)
			}

			logger.Infof("Requested IPA patch for \"%s\"", ipa)

			extract()
			loadInfo()

			setReactNavigationName()
			setSupportedDevices()
			setFileAccess()
			setAppName()
			setIcons()

			saveInfo()
			archive()

			exit()
			return nil
		},
	}

	assets = os.TempDir()

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}

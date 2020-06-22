package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/vburenin/go-vphoto/downloader"
	"github.com/vburenin/go-vphoto/handlers"
	"github.com/vburenin/go-vphoto/settings"
	"net/http"
	"os"
	"path"

	"github.com/urfave/cli/v2"
)

func main() {


	app := &cli.App{
		Name: "Photo frame",
		Usage: "A photoframe service",
		Action: func(c *cli.Context) error {
			return nil
		},
		Flags: []cli.Flag {
			&cli.StringFlag{
				Name: "cache",
				Value: "cache",
				Usage: "Downloaded pictures cache location",
				Destination: &settings.CacheLocation,
				EnvVars: []string{"VPHOTO_CACHE"},
			},
			&cli.StringFlag{
				Name: "workdir",
				Value: ".",
				Usage: "Downloaded pictures cache location",
				Destination: &settings.WorkDir,
				EnvVars: []string{"VPHOTO_WORKDIR"},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	settings.InitSettings()
	downloader.InitDownloader()

	staticFilePath := path.Join(settings.WorkDir, "static")
	log.Infof("static file location: %s", staticFilePath)
	log.Infof("file cache location: %s", settings.CacheLocation)

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir(staticFilePath))))

	http.Handle("/cache/",
		http.StripPrefix("/cache/",
			http.FileServer(http.Dir(settings.CacheLocation))))

	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/callback", handlers.GoogleCallBackHandler)
	http.HandleFunc("/albums", handlers.ListAlbums)
	http.HandleFunc("/pictures", handlers.SlideShow)
	http.HandleFunc("/loadpic", handlers.LoadPicture)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

}

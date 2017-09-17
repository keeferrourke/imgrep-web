package main

import (
	/* Standard library packages */

	/* Third party */
	// imports as "cli", pinned to v1; cliv2 is going to be drastically
	// different and pinning to v1 avoids issues with unstable API changes
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	cli "gopkg.in/urfave/cli.v1"

	/* Local packages */
	"github.com/gorilla/mux"
	"github.com/keeferrourke/imgrep/files"
	"github.com/keeferrourke/imgrep/storage"
)

var PORT string

type ResultRow struct {
	Filename string `json:"filename"`
	Bytes    []byte `json:"bytes"`
}

var Server = cli.Command{
	Name:    "run",
	Aliases: []string{"start"},
	Usage:   "initialize gui server",
	Action:  StartServer,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "port, p",
			Value:       "1337",
			Usage:       "configure port",
			Destination: &PORT,
		},
	},
}

func StartServer(c *cli.Context) error {
	r := mux.NewRouter()
	go files.InitFromPath(c)
	r.HandleFunc("/imgrep/search", func(w http.ResponseWriter, r *http.Request) {
		keyword := r.FormValue("keyword")
		keyword = strings.TrimSpace(keyword)

		keywordList := strings.Split(keyword, " ")

		results := []*ResultRow{}
		for _, kw := range keywordList {
			filenames, err := storage.Get(kw)

			if err != nil {
				log.Fatalf(err.Error())
			}

			for _, file := range filenames {
				found := false

				for _, rr := range results {
					if rr.Filename == file {
						found = true
					}
				}

				if found {
					continue
				}

				f, err := ioutil.ReadFile(file)
				if err != nil {
					log.Fatalf(err.Error())
				}

				results = append(results, &ResultRow{
					Filename: file,
					Bytes:    f,
				})
			}

		}

		resp := map[string][]*ResultRow{}
		resp["files"] = results

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates, err := template.ParseFiles("./index.html")
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	s := http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets")))
	r.PathPrefix("/assets/").Handler(s)
	fmt.Println("Started server on localhost:" + PORT)

	http.ListenAndServe(":"+PORT, r)

	return nil
}

func main() {
	// customize cli
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer, "%s %s - %s\n",
			c.App.Name, c.App.Version, c.App.Description)
	}

	// set up the application
	app := cli.NewApp()
	app.Authors = []cli.Author{
		cli.Author{
			Name: "Keefer Rourke",
		},
		cli.Author{
			Name: "Ivan Zhang",
		},
	}
	app.Copyright = "(c) 2017 under the MIT License"
	app.EnableBashCompletion = true
	app.Name = "imgrep-web"
	app.Description = "web interface for using imgrep"
	app.Usage = "grep image files for words"
	app.Version = "v0"
	app.Commands = []cli.Command{
		Server,
	}
	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.Writer, "Did you read the manual?\n")
	}

	//connect to local sqlite
	storage.InitDB(files.DBFILE)

	app.Run(os.Args)
	return
}

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

var (
	PORT   string
	TPLDIR string
	ASSETS string
)

type ResultRow struct {
	Filename string `json:"filename"`
	Bytes    []byte `json:"bytes"`
}

// reliably serve files even when binary is installed and run from arbitrary
// working directory
// this is horrible though so try to find a better way to this? :(
func setPath(dir string) string {
	basePath := "src" + string(os.PathSeparator) + "github.com" + string(os.PathSeparator) + "keeferrourke" + string(os.PathSeparator) + "imgrep-web"
	p := os.Getenv("GOPATH") + string(os.PathSeparator) + basePath + string(os.PathSeparator) + dir
	if _, err := os.Stat(p); os.IsNotExist(err) {
		p = os.Getenv("GOROOT") + string(os.PathSeparator) + dir
		if _, err := os.Stat(p); os.IsNotExist(err) {
			log.Fatal("path error: can not stat directory " + p)
		}
	}
	return p
}

func StartServer(c *cli.Context) error {
	r := mux.NewRouter()

	go files.InitFromPath(false)

	r.HandleFunc("/imgrep/search", func(w http.ResponseWriter, r *http.Request) {
		keyword := r.FormValue("keyword")
		keyword = strings.TrimSpace(keyword)

		keywordList := strings.Split(keyword, " ")

		results := []*ResultRow{}
		for _, kw := range keywordList {
			filenames, err := storage.Get(kw, true) // case insensitive search

			if err != nil {
				log.Printf(err.Error())
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

				// remove non-existing files from database
				if _, err := os.Stat(file); os.IsNotExist(err) {
					storage.Delete(file)
					continue
				}

				f, err := ioutil.ReadFile(file)
				if err != nil {
					log.Printf(err.Error())
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
		index := TPLDIR + string(os.PathSeparator) + "index.html"
		templates, err := template.ParseFiles(index)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	s := http.StripPrefix("/assets/", http.FileServer(http.Dir(ASSETS)))
	r.PathPrefix("/assets").Handler(s)
	fmt.Println("Started server on localhost:" + PORT)

	http.ListenAndServe(":"+PORT, r)

	return nil
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

func init() {
	TPLDIR = setPath("tpl")
	ASSETS = setPath("assets")
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
			Name:  "Keefer Rourke",
			Email: "mail@krourke.org",
		},
		cli.Author{
			Name:  "Ivan Zhang",
			Email: "ivan@ivanzhang.ca",
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

	// connect to local sqlite
	storage.InitDB(files.DBFILE)

	app.Run(os.Args)
	return
}

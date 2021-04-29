package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/rif/telegraf-freeswitch/utils"
)

type Config struct {
	Host 	string `default:"localhost"`
	Port	int `default:"8021"`
	Pass 	string `default:"ClueCon"`
	Serve 	bool `default:"false"`
	Execd	bool `default:"false"`
	ListenAddress string `default:"127.0.0.1"`
	ListenPort int `default:"9191"`
}

func handler(w http.ResponseWriter, route string) {
}

func main() {

	l := log.New(os.Stderr, "", 0)

	var c Config
	err := envconfig.Process("telegraf_freeswitch", &c)
	if err != nil {
		l.Fatal("error parsing config: ", err)
	}

	fetcher, err := utils.NewFetcher(c.Host, c.Port, c.Pass)
	if err != nil {
		l.Fatal("error connecting to fs: ", err)
	}

	defer fetcher.Close()
	if !c.Serve {
		if c.Execd {
			reader := bufio.NewReader(os.Stdin)
			for {
				text, err := reader.ReadString('\n')
				if err != nil {
					l.Print("error reading from stdin: ", err)
					continue
				}
				if strings.TrimSpace(text) != "" {
					break
				}
				if err := fetcher.GetData(); err != nil {
					l.Print(err.Error())
				}
				fmt.Print(fetcher.FormatOutput(utils.InfluxFormat))
			}
			os.Exit(0)
		}
		if err := fetcher.GetData(); err != nil {
			l.Print(err.Error())
		}
		fmt.Print(fetcher.FormatOutput(utils.InfluxFormat))
		os.Exit(0)
	}

	http.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		if err := fetcher.GetData(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		status, _ := fetcher.FormatOutput(utils.JSONFormat)
		if _, err := w.Write([]byte(status)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/profiles/", func(w http.ResponseWriter, r *http.Request) {
		if err := fetcher.GetData(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, profiles := fetcher.FormatOutput(utils.JSONFormat)
		if _, err := w.Write([]byte(profiles)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	listen := fmt.Sprintf("%s:%d", c.ListenAddress, c.ListenPort)
	fmt.Printf("Listening on %s...", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}

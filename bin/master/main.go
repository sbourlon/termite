package main

import (
	"flag"
	"log"
	"github.com/hanwen/termite/termite"
	"io/ioutil"
	"strings"
)

func main() {
	cachedir := flag.String("cachedir", "/tmp/fsserver-cache", "content cache")
	workers := flag.String("workers", "", "comma separated list of worker addresses")
	coordinator := flag.String("coordinator", "localhost:1233",
		"address of coordinator. Overrides -workers")
	socket := flag.String("socket", ".termite-socket", "socket to listen for commands")
	exclude := flag.String("exclude", "/sys,/proc,/dev,/selinux,/cgroup", "prefixes to not export.")
	secretFile := flag.String("secret", "/tmp/secret.txt", "file containing password.")
	jobs := flag.Int("jobs", 1, "number of jobs to run")

	flag.Parse()
	secret, err := ioutil.ReadFile(*secretFile)
	if err != nil {
		log.Fatal("ReadFile", err)
	}

	workerList := strings.Split(*workers, ",")
	excludeList := strings.Split(*exclude, ",")
	c := termite.NewContentCache(*cachedir)
	master := termite.NewMaster(
		c, *coordinator, workerList, secret, excludeList, *jobs)
	master.Start(*socket)
}

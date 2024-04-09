package main

import (
	"flag"
	"github.com/isongjosiah/hack/tar/gotar"
	"log"
	"os"
)

var tarFlag = gotar.TarFlag{}
var tarFlags = flag.NewFlagSet("gotar", flag.ExitOnError)

func init() {

	// define the flags
	tarFlags.BoolVar(&tarFlag.ListArchivedContent, "t", false, "list archive contents to stdout.")

}

func main() {

	// parse the flags
	if err := tarFlags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
		return

	}

	gotar.IgniteTarEngine(tarFlag, os.Args).Execute()

}

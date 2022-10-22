package main

import (
	"flag"
	"log"
	"os"

	"github.com/ivanilves/travelgrunt/pkg/directory"
	"github.com/ivanilves/travelgrunt/pkg/directory/tree"
	"github.com/ivanilves/travelgrunt/pkg/file"
	"github.com/ivanilves/travelgrunt/pkg/filter"
	"github.com/ivanilves/travelgrunt/pkg/menu"
	"github.com/ivanilves/travelgrunt/pkg/scm"
	"github.com/ivanilves/travelgrunt/pkg/shell"
	"github.com/ivanilves/travelgrunt/pkg/terminal"
)

var appVersion = "default"

var outFile string
var top bool
var version bool

func init() {
	flag.StringVar(&outFile, "out-file", "", "output project path into the file specified instead of spawning a shell")
	flag.BoolVar(&top, "top", false, "get to the repository top level (root) path and exit")
	flag.BoolVar(&version, "version", false, "print application version and exit")
}

func usage() {
	println("Usage: " + shell.Name() + " [<match> <match2> ... <matchN>]")
	println("")
	println("Options:")
	flag.PrintDefaults()
}

func writeFileAndExit(fileName string, data string) {
	if err := file.Write(fileName, data); err != nil {
		log.Fatalf("failed to write file (%s): %s", fileName, err.Error())
	}

	os.Exit(0)
}

func main() {
	if shell.IsRunningInside() {
		log.Fatalf("%s already running (pid: %d), please type \"exit\" to return to the parent shell first", shell.Name(), shell.Getppid())
	}

	flag.Usage = usage
	flag.Parse()

	if version {
		println(appVersion)

		os.Exit(0)
	}

	matches := flag.Args()

	rootPath, err := scm.RootPath()

	if err != nil {
		log.Fatalf("failed to extract top level filesystem path from SCM: %s", err.Error())
	}

	if top {
		writeFileAndExit(outFile, rootPath)
	}

	entries, paths, err := directory.Collect(rootPath)

	if err != nil {
		log.Fatalf("failed to collect Terragrunt project directories: %s", err.Error())
	}

	if err := filter.Validate(matches); err != nil {
		log.Fatalf("invalid filter: %s", err.Error())
	}

	paths = filter.Apply(paths, matches)

	tree := tree.NewTree(paths)

	var selected string
	var parentID string

	for c := -1; c < tree.LevelCount()-1; c++ {
		if !tree.HasChildren(c, parentID) {
			break
		}

		selected, err = menu.Build(tree.ChildNames(c, parentID), terminal.Height())
		if err != nil {
			log.Fatalf("failed to build menu: %s", err.Error())
		}

		parentID = tree.ChildItems(c, parentID)[selected]

		selected = parentID
	}

	if outFile != "" {
		writeFileAndExit(outFile, entries[selected])
	}

	shell.Spawn(entries[selected])
}
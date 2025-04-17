package main

// Package flag implements command-line flag parsing.
import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func scanGitFolders(folders []string, folder string) []string {
	/*
	****  we get a slice of strings from recursiveScanFolder()
	****	we get the path of the dot file we’re going to write to.
	****	we write the slice contents to the file
	 */
	folder = strings.TrimSuffix(folder, "/") // if the end of folder has "/"  removes and returns the string else returns unchanged

	f, err := os.Open(folder)

	if err != nil {
		log.Fatal("error while opening the file: ", err)
	}

	files, err := f.Readdir(-1)
	f.Close()

	if err != nil {
		log.Fatal(err)
	}

	var path string

	for _, file := range files {

		if file.IsDir() {
			path = folder + "/" + file.Name()

			if file.Name() == ".git" {
				path = strings.TrimSuffix(path, "/.git")

				fmt.Println("path:", path)

				folders = append(folders, path)

				continue
			}

			if file.Name() == "vendor" || file.Name() == "node_modules" {
				continue
			}
			folders = scanGitFolders(folders, path)
		}
	}

	return folders
}

func recursiveScanGitFolders(folder string) []string {
	return scanGitFolders(make([]string, 0), folder)
}

func scan(folder string) {

	/*
		====================
			scanning a folder
		====================

		ACCUIRING A LIST OF FOLDERS TO SCAN:
		The algorithm ill follow for this first part is pretty simple:

		*	pick a directory location
								 |
								\/
		*	scan for .git folders in there
			and all the sub directories
								 |
								\/
		*	Make a slice of folder
			paths containing .git
								|
							 \/
		*	Store the repo paths in
			~/.gogitlocalstats,
			on per line
	*/

	repositories := recursiveScanGitFolders(folder)

	print("scan")
	print("scan", repositories)
}

func stats(email string) {
	print("stats")
}

func main() {
	var folder string
	var email string

	// If you like, you can bind the flag to a variable using the Var() functions.
	// flag.IntVar(&flagvar, "flagname", 1234, "help message for flagname")

	// 	After all flags are defined, call
	// 	flag.Parse()
	// to parse the command line into the defined flags.

	flag.StringVar(&folder, "add", "", "add a new folder to scan for Git repositories")
	flag.StringVar(&email, "email", "your@email.com", "the email to scan")

	flag.Parse()

	// if the [-add "/some/directorys"] is used
	if folder != "" {
		scan(folder)
		return
	}

	stats(email)
}

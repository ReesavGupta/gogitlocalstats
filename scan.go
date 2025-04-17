package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"slices"
	"strings"
)

// Package flag implements command-line flag parsing.
// dumpStringsSliceToFile writes content to the file in path `filePath` (overwriting existing content)
func dumpStringsSliceToFile(repos []string, filePath string) {
	content := strings.Join(repos, "\n")
	os.WriteFile(filePath, []byte(content), 0755)
}

func sliceContains(slice []string, value string) bool {

	/*
		same thing

			for _, val := range slice {
				if val == value {
					return true
				}
			}
			return false
	*/
	return slices.Contains(slice, value)
}

func joinRepos(existingRepos []string, newRepos []string) []string {

	for _, repo := range newRepos {
		if !sliceContains(existingRepos, repo) {
			existingRepos = append(existingRepos, repo)
		}
	}

	return existingRepos

}

func openFile(filePath string) *os.File {
	// opens the file in append + write-only mode.
	// the third argn is - owner has read/write/execute, group && others has read/write
	// f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0777) //was giving error
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			_, err = os.Create(filePath)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err) // some other error
		}
	}

	return f

}

// parseFileLinesToSlice given a file path string, gets the content of each line and parses it to a slice of strings.
func parseFileLinesToSlice(filePath string) []string {
	f := openFile(filePath)

	defer f.Close()

	var lines []string

	scanner := bufio.NewScanner(f)

	for scanner.Scan() { //returns false when there are no more tokens to scan
		lines = append(lines, scanner.Text())
	}

	//After the loop, check if the scanner had any error.
	//	If the error is not io.EOF, panic.
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	return lines

}

// addNewSliceElementsToFile given a slice of strings representing paths, stores them
// to the filesystem
func addNewSliceElementsToFile(filePath string, newRepos []string) {
	existingRepos := parseFileLinesToSlice(filePath)
	allRepos := joinRepos(existingRepos, newRepos)
	dumpStringsSliceToFile(allRepos, filePath)
}

// getDotFilePath returns the dot file for the respos list.
// creates it and the enclosing folder if it does not exist.

// This function uses the os/user package’s Current function to get the current user, which is a struct defined as

/*
// User represents a user account.

	type User struct {
		// Uid is the user ID.
		// On POSIX systems, this is a decimal number representing the uid.
		// On Windows, this is a security identifier (SID) in a string format.
		// On Plan 9, this is the contents of /dev/user.

		Uid string

		// Gid is the primary group ID.
		// On POSIX systems, this is a decimal number representing the gid.
		// On Windows, this is a SID in a string format.
		// On Plan 9, this is the contents of /dev/user.

		Gid string

		// Username is the login name.

		Username string

		// Name is the user's real or display name.
		// It might be blank.
		// On POSIX systems, this is the first (or only) entry in the GECOS field
		// list.
		// On Windows, this is the user's display name.
		// On Plan 9, this is the contents of /dev/user.

		Name string

		// HomeDir is the path to the user's home directory (if they have one).

		HomeDir string

	}
*/
func getDotFilePath() string {
	u, err := user.Current()

	if err != nil {
		log.Fatal(err)
	}

	dotFile := u.HomeDir + `\.gogitlocalstats`

	println("\nthis is the dotFilePath:", dotFile, "\n")
	return dotFile

	// So, now we have a list of repos, a file to write them to, and the next step for scan() is to store them, without adding duplicate lines.
}

func scanGitFolders(folders []string, folder string) []string {
	/*
	**** 1. we get a slice of strings from recursiveScanFolder()
	**** 2. we get the path of the dot file we’re going to write to.
	**** 3.	we write the slice contents to the file
	 */
	folder = strings.TrimSuffix(folder, "/") // if the end of folder has "/"  removes and returns the string else returns unchanged

	f, err := os.Open(folder)

	if err != nil {
		log.Fatal("error while opening the folder: ", err)
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
	dotFilePath := getDotFilePath()
	addNewSliceElementsToFile(dotFilePath, repositories)
}

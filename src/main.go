package main

import "flag"

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

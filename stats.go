package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// yeah need to work over here as well --> we will be generating the stats over here
/*
	============================
				CALCULATE STATS
	============================
*/

/*
	read the repositories list
					 |
					\/
	i = repositories no. count
					 |

else				\/
------ if i < count <-------------------------
|						|																|
|					 \/																|
|		calculate contribution per							|
|		day for the last 52 weeks --------------|
|
|else
|else
|else
---------> render contribution graph
*/
const outOfRange = 99999
const daysInLastSixMonths = 183
const weeksInLastSixMonths = 26

type column []int

// printCell given a cell value prints it with a different format
// based on the value amount, and on the `today` flag.
func printCell(val int, today bool) {
	escape := "\033[0;37;30m"
	switch {
	case val > 0 && val < 5:
		escape = "\033[1;30;47m"
	case val >= 5 && val < 10:
		escape = "\033[1;30;43m"
	case val >= 10:
		escape = "\033[1;30;42m"
	}

	if today {
		escape = "\033[1;37;45m"
	}

	if val == 0 {
		fmt.Printf("%s  - %s", escape, "\033[0m")
		return
	}

	str := "  %d "
	switch {
	case val >= 10:
		str = " %d "
	case val >= 100:
		str = "%d "
	}

	fmt.Printf(escape+str+"\033[0m", val)
}

// printDayCol given the day number (0 is Sunday) prints the day name,
// alternating the rows (prints just 2,4,6)
func printDayCol(day int) {
	out := "     "
	switch day {
	case 1:
		out = " Mon "
	case 3:
		out = " Wed "
	case 5:
		out = " Fri "
	}

	fmt.Print(out)
}

// printMonths prints the month names in the first line, determining when the month
// changed between switching weeks
func printMonths() {
	week := getBeginningOfDay(time.Now()).Add(-(daysInLastSixMonths * time.Hour * 24))
	month := week.Month()
	fmt.Printf("         ")
	for {
		if week.Month() != month {
			fmt.Printf("%s ", week.Month().String()[:3])
			month = week.Month()
		} else {
			fmt.Printf("    ")
		}

		week = week.Add(7 * time.Hour * 24)
		if week.After(time.Now()) {
			break
		}
	}
	fmt.Printf("\n")
}

// printCells prints the cells of the graph
func printCells(cols map[int]column) {
	printMonths()

	today := time.Now()
	todayWeek := countDaysSinceDate(today) / 7
	todayDayInWeek := int(today.Weekday())

	for j := 6; j >= 0; j-- {
		for i := weeksInLastSixMonths + 1; i >= 0; i-- {
			if i == weeksInLastSixMonths+1 {
				printDayCol(j)
			}

			if col, ok := cols[i]; ok {
				// Check if this is today's cell
				isToday := (i == todayWeek && j == todayDayInWeek)

				if len(col) > j {
					printCell(col[j], isToday)
					continue
				}
			}
			printCell(0, false)
		}
		fmt.Printf("\n")
	}
}

// buildCols generates a map with rows and columns ready to be printed to screen
func buildCols(keys []int, commits map[int]int) map[int]column {
	cols := make(map[int]column)

	for _, day := range keys {
		if day >= daysInLastSixMonths {
			continue
		}

		// Calculate position in the grid
		weekNum := day / 7
		dayInWeek := day % 7

		// Initialize column if needed
		if _, exists := cols[weekNum]; !exists {
			cols[weekNum] = make(column, 7)
		}

		// Set commit count
		cols[weekNum][dayInWeek] = commits[day]
	}

	return cols
}

// sortMapIntoSlice returns a slice of indexes of a map, ordered
func sortMapIntoSlice(m map[int]int) []int {
	// order map
	// To store the keys in slice in sorted order
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	return keys
}

// printCommitsStats prints the commits stats
func printCommitsStats(commits map[int]int) {
	keys := sortMapIntoSlice(commits)
	cols := buildCols(keys, commits)
	printCells(cols)
}

// getBeginningOfDay given a time.Time calculates the start time of that day
func getBeginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return startOfDay
}

// countDaysSinceDate counts how many days passed since the passed `date`
func countDaysSinceDate(date time.Time) int {
	now := getBeginningOfDay(time.Now())
	date = getBeginningOfDay(date)

	if date.After(now) {
		return outOfRange
	}

	duration := now.Sub(date)
	days := int(duration.Hours() / 24)

	if days > daysInLastSixMonths {
		return outOfRange
	}

	return days
}

// calcOffset determines and returns the amount of days missing to fill
// the last row of the stats graph
func calcOffset() int {
	var offset int
	weekday := time.Now().Weekday()

	switch weekday {
	case time.Sunday:
		offset = 7
	case time.Monday:
		offset = 6
	case time.Tuesday:
		offset = 5
	case time.Wednesday:
		offset = 4
	case time.Thursday:
		offset = 3
	case time.Friday:
		offset = 2
	case time.Saturday:
		offset = 1
	}

	return offset
}
func fillCommits(email string, path string, commits map[int]int) map[int]int {
	repo, err := git.PlainOpen(path)
	if err != nil {
		fmt.Printf("Error opening repo %s: %v\n", path, err)
		return commits
	}

	repoHeadRef, err := repo.Head()
	if err != nil {
		fmt.Printf("Skipping repo %s due to error: %v\n", path, err)
		return commits
	}
	fmt.Printf("HEAD of %s points to: %s (%s)\n", path, repoHeadRef.Name(), repoHeadRef.Hash())

	iterator, err := repo.Log(&git.LogOptions{From: repoHeadRef.Hash()})
	if err != nil {
		fmt.Printf("Skipping log reading for repo %s: %v\n", path, err)
		return commits
	}
	iterator.ForEach(func(c *object.Commit) error {
		fmt.Printf("[%s] %s by %s <%s>\n", path, c.Committer.When.Format("2006-01-02"), c.Author.Name, c.Author.Email)
		return nil
	})

	// Reset iterator for the actual processing
	iterator, _ = repo.Log(&git.LogOptions{From: repoHeadRef.Hash()})

	err = iterator.ForEach(func(c *object.Commit) error {
		if c.Author.Email != email {
			return nil
		}

		daysAgo := countDaysSinceDate(c.Author.When)

		if daysAgo != outOfRange {
			commits[daysAgo]++
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error iterating commits for %s: %v\n", path, err)
	}

	return commits
}

func processRepositories(email string) map[int]int {
	filePath := getDotFilePath()
	repos := parseFileLinesToSlice(filePath)

	commits := make(map[int]int, daysInLastSixMonths)

	// Initialize all days with zero commits
	for i := range daysInLastSixMonths {
		commits[i] = 0
	}

	for _, path := range repos {
		commits = fillCommits(email, path, commits)
	}
	return commits
}

func stats(email string) {
	print("stats")

	commits := processRepositories(email)
	printCommitsStats(commits)
}

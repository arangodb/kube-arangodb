package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const Root util.EnvironmentVariable = "ROOT"

func main() {
	if err := mainE(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

type dateRange struct {
	from, to int
}

func mainE() error {
	// Ensure that all files have proper license dates

	rewrite := false

	for _, a := range os.Args[1:] {
		if a == "-w" {
			rewrite = true
		}
	}

	files := map[string]int{}

	currentHeaders := map[string]dateRange{}

	// Extract current dates
	for _, file := range os.Args[1:] {
		if file == "-w" {
			continue
		}

		var out bytes.Buffer

		cmd := exec.Command("git", "log", "-n", "1", "--pretty=format:%cd", file)

		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "Unable to parse file %s", file)
		}

		c := out.String()

		if c == "" {
			continue
		}

		d, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", c)
		if err != nil {
			return errors.Wrapf(err, "Unable to parse file %s", file)
		}

		files[file] = d.Year()
	}

	// Extract license dates
	for file := range files {
		if from, to, err := extractFileLicenseData(file); err != nil {
			return errors.Wrapf(err, "Unable to parse file %s", file)
		} else {
			currentHeaders[file] = dateRange{
				from: from,
				to:   to,
			}
		}
	}

	valid := true

	for file, date := range files {
		if c, ok := currentHeaders[file]; !ok {
			println(fmt.Sprintf("Date not discovered for %s", file))
			valid = false
		} else if date < c.from || date > c.to {
			println(fmt.Sprintf("Date %d not in range %d-%d for %s. File has beed modified", date, c.from, c.to, file))
			if rewrite {
				println("Rewrite file")

				q := fmt.Sprintf("// Copyright %d-%d ArangoDB GmbH, Cologne, Germany", c.from, c.to)
				if c.from == c.to {
					q = fmt.Sprintf("// Copyright %d ArangoDB GmbH, Cologne, Germany", c.to)
				}

				changed, err := rewriteLicenseDates(file, q, fmt.Sprintf("// Copyright %d-%d ArangoDB GmbH, Cologne, Germany", c.from, date))
				if err != nil {
					return err
				} else if changed {
					continue
				}
			}
			valid = false
		}
	}

	if !valid {
		return errors.Newf("Parse of file failed")
	}

	return nil
}

func rewriteLicenseDates(file string, from, to string) (bool, error) {
	data, changed, err := readNewLicenseDates(file, from, to)
	if err != nil {
		return false, err
	}

	if !changed {
		return false, nil
	}

	if err := os.WriteFile(file, data, 0644); err != nil {
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func readNewLicenseDates(file string, from, to string) ([]byte, bool, error) {
	readFile, err := os.Open(file)
	if err != nil {
		return nil, false, err
	}

	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	q := bytes.NewBuffer(nil)

	got := false

	for fileScanner.Scan() {
		t := fileScanner.Text()
		if t == from {
			got = true
			q.WriteString(to)
		} else {
			q.WriteString(t)
		}
		q.WriteString("\n")
	}

	return q.Bytes(), got, nil
}

func extractFileLicenseData(file string) (int, int, error) {
	readFile, err := os.Open(file)
	if err != nil {
		return 0, 0, err
	}

	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		t := fileScanner.Text()

		if !strings.HasPrefix(t, "// Copyright ") {
			continue
		}
		t = strings.TrimPrefix(t, "// Copyright ")

		if !strings.HasSuffix(t, " ArangoDB GmbH, Cologne, Germany") {
			continue
		}
		t = strings.TrimSuffix(t, " ArangoDB GmbH, Cologne, Germany")

		if !strings.Contains(t, "-") {
			t = fmt.Sprintf("%s-%s", t, t)
		}

		n := strings.Split(t, "-")

		from, err := strconv.Atoi(n[0])
		if err != nil {
			return 0, 0, err
		}

		to, err := strconv.Atoi(n[1])
		if err != nil {
			return 0, 0, err
		}

		return from, to, nil
	}

	return 0, 0, errors.Newf("Unable to find license string")
}

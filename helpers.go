package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// TOOD - this was pulled from rebate service verbatum - remove to decrease chance of bugs and duplicate code
// this only works for single and double column CSVs
func convertCSVToMap(fileName string, swap bool, debug bool) map[string]string {
	if debug {
		fmt.Println("Reading in dependency data")
	}
	basePath := ""

	result := make(map[string]string)
	file, err := os.Open(basePath + fileName)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if debug { fmt.Println(line) }

		parts := strings.Split(line, ",")
		if (len(parts) > 1) {
			if swap {
				result[parts[1]] = parts[0]
			} else {
				result[parts[0]] = parts[1]
			}
		} else {
			result[parts[0]] = ""
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	if debug {
		fmt.Println(result)
		fmt.Println("")
		fmt.Println("Successfully read all dependency data!")
	}

	return result
}

// TODO - this was pulled from rebate service - make sure to reuse that code once its foud a home
func parseMultiColcsv(fileName string, debug bool) []map[string]string {
	if debug {
		fmt.Println("Reading in csr data")
	}
	basePath := ""

	var result []map[string]string

	file, err := os.Open(basePath + fileName)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	headerSet := false
	var keys []string
	for scanner.Scan() {
		line := scanner.Text()
		if debug {
			fmt.Println(line)
		}

		// location must be consistent for this to work
		if !headerSet {
			headers := strings.Split(line, ",")
			keys = headers
			headerSet = true
		} else {
			values := strings.Split(line, ",")
			data := make(map[string]string)
			for index, key := range keys {
				data[key] = values[index]
				// make sure this order is the same every time
			}
			result = append(result, data)
			if debug {
				fmt.Println(data)
				fmt.Println()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	if debug {
		fmt.Println(result)
		fmt.Println("")
		fmt.Println("Successfully read all dependency data!")
	}

	return result
}
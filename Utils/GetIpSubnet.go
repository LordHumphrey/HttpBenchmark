package Utils

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func GetIpSubnetFromFile(fileFullPath string, lineNum int) ([]string, error) {
	file, err := os.Open(fileFullPath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Errorf("error closing file: %v", err)
		}
	}(file)

	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	scanner := bufio.NewScanner(file)
	var lines []string
	var count int
	for scanner.Scan() {
		count++
		line := scanner.Text()
		if len(lines) < lineNum {
			lines = append(lines, line)
		} else {
			// With probability lineNum/count, replace a random line in lines with the current line.
			if r.Float64() < float64(lineNum)/float64(count) {
				lines[r.Intn(lineNum)] = line
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return lines, nil
}

func GetIpSubnetFromEmbedFile(data string, lineNum int) ([]string, error) {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	scanner := bufio.NewScanner(strings.NewReader(data))
	var lines []string
	var count int
	for scanner.Scan() {
		count++
		line := scanner.Text()
		if len(lines) < lineNum {
			lines = append(lines, line)
		} else {
			// With probability lineNum/count, replace a random line in lines with the current line.
			if r.Float64() < float64(lineNum)/float64(count) {
				lines[r.Intn(lineNum)] = line
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return lines, nil
}

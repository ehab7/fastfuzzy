package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ehab7/fastfuzzy/algo"
	"github.com/ehab7/fastfuzzy/configure"
)

var debug = false
var appConfig = &configure.Conf{}

func init() {
	var err error
	if appConfig, err = appConfig.GetConfig(); err != nil {
		log.Fatalln(err)
	}

}

func ProcessFile(config *configure.Conf) (bool, error) {
	var (
		inFile *os.File
		err    error
	)

	if config.InputFile != "" {
		inFile, err = os.Open(config.InputFile)
		if err != nil {
			return false, err
		}
		defer inFile.Close()
	} else {
		inFile = os.Stdin
	}

	var pos = config.Position
	debug = config.Debug
	sc := bufio.NewScanner(inFile)

	for sc.Scan() {
		csvReader := sc.Text()
		var line = csvReader

		if config.Separator != "" {
			force := strings.Split(csvReader, config.Separator)
			if (len(force) == 0) || (len(force) <= pos) {
				continue
			}
			line = strings.Trim(force[pos], " ")
		}
		// bypass  short words 2 or less char
		if len(line) <= 2 {
			continue
		}
		line = strings.ToLower(line)

		// clean up word by removing unwanted phrases/chars
		for _, removeWord := range config.Remove {
			line = strings.Replace(line, removeWord, "", -1)
		}
		// bypass rejected  words
		var breakMainBlock = false
		for _, rejectWord := range config.Reject {
			if strings.Contains(line, rejectWord) {
				breakMainBlock = true
				break
			}
		}

		if breakMainBlock {
			continue
		}

		found, total := algo.Process(&line, &appConfig.Algo)
		if found {
			if total > 1.0 && debug {
				log.Printf("%s -> fuzzyDebug: direct match\n", csvReader)
			} else if debug {
				log.Printf("%s -> fuzzyDebug: algo score %f\n", csvReader, total)
			}
			fmt.Println(csvReader)
		} else if total > 0.5 {
			// check for the included keywords.
			for _, includeWord := range config.Include {
				if strings.Contains(line, includeWord) {
					if debug {
						log.Printf("%s -> fuzzy score %f < min or soundex is False forced by inclusion\n", csvReader, total)
					}
					fmt.Println(csvReader)
					break
				}
			}
		}

	}
	return true, nil
}

func main() {

	if _, err := ProcessFile(appConfig); err != nil {
		log.Fatalln("failed to run:", err.Error())

	}

}

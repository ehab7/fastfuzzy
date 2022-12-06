package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	//"fmt"
	"sync"

	"github.com/ehab7/fastfuzzy/algo"
	"github.com/ehab7/fastfuzzy/configure"

	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

var debug = false
var appConfig = &configure.Conf{}
var alogNodes []algo.AlgoUnit

type ResultsToWrite struct {
	Line   string `parquet:"name=line, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Result string `parquet:"name=result, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

type item struct {
	line   *string
	lineno int
}

type itemResult struct {
	line   *string
	lineno int
	score  float32
	name   *string
}

func init() {
	var err error
	if appConfig, err = appConfig.GetConfig(); err != nil {
		log.Fatalln(err)
	}
	for i := 0; i < len(appConfig.Nodes); i++ {
		alogNodes = append(alogNodes, algo.InitAlgo(appConfig.Nodes[i].Name, appConfig.Nodes[i].FuzzySearch, appConfig.Nodes[i].Threshold))
	}
}

func processingNode(wg *sync.WaitGroup, feedIn chan item, feedOut chan itemResult, index int) {
	defer wg.Done()
	x := itemResult{}
	x.name = &(alogNodes[index].Name)
	nodeConfig := appConfig.Nodes[index]
	skip := false
	//log.Println("listing to channel ", index)
	for {
		c := <-feedIn
		x.line = c.line
		x.lineno = c.lineno
		x.score = 0.0

		skip = false
		for i := 0; i < len(nodeConfig.IgnoreIfHas); i++ {
			if strings.Contains(*c.line, nodeConfig.IgnoreIfHas[i]) {
				skip = true
				break
			}
		}

		if !skip && len(nodeConfig.MustHas) > 0 {
			skip = true
			for i := 0; i < len(nodeConfig.MustHas); i++ {
				if strings.Contains(*c.line, nodeConfig.MustHas[i]) {
					skip = false
					break
				}
			}

		}

		if !skip {
			for i := 0; i < len(nodeConfig.ExplicitSearch); i++ {
				if strings.Contains(*c.line, nodeConfig.ExplicitSearch[i]) {
					x.score = 1.0
					skip = true
					break
				}
			}
		}

		if !skip {
			for i := 0; i < len(nodeConfig.ExplicitSearchCombine); i++ {
				allFound := true
				for k := 0; k < len(nodeConfig.ExplicitSearchCombine[i]); k++ {
					if !strings.Contains(*c.line, nodeConfig.ExplicitSearchCombine[i][k]) {
						allFound = false
						break
					}
					if allFound {
						skip = true
						x.score = 1.0
					}
				}
			}
		}

		if !skip {
			_, score := algo.ProcessSentanceFuzzy(c.line, &alogNodes[index])
			x.score = score

		}

		//if !skip {
		feedOut <- x
		//}

		break
	}

}

func ProcessFile(toNodes []chan item, fromNodes chan itemResult) (bool, error) {
	var (
		inFile        *os.File
		err           error
		skip          bool
		counter       int
		maxScoreName  string
		maxScoreValue float32
	)

	inFile, err = os.Open(appConfig.InputFile)
	if err != nil {
		return false, err
	}
	defer inFile.Close()

	var pw *writer.ParquetWriter
	var pos = appConfig.Position
	var outputToFile = false

	if appConfig.OutputFile != "" {
		outputToFile = true
		w, err := os.Create("output/flat.parquet")
		if err != nil {
			log.Println("Can't create local file", err)
			return false, err
		}

		pw, err = writer.NewParquetWriterFromWriter(w, new(ResultsToWrite), 4)
		if err != nil {
			log.Println("Can't create parquet writer", err)
			return false, err
		}
		pw.RowGroupSize = 128 * 1024 * 1024 //128M
		pw.CompressionType = parquet.CompressionCodec_SNAPPY
	}

	sc := bufio.NewScanner(inFile)
	var wg sync.WaitGroup
	counter = 0
	for sc.Scan() {
		maxScoreName = ""
		maxScoreValue = 0.1
		skip = false

		csvReader := sc.Text()
		var got = csvReader
		var line = csvReader
		line = strings.ToLower(line)

		for i := 0; i < len(appConfig.Nodes); i++ {
			wg.Add(1)
			i := i
			go processingNode(&wg, toNodes[i], fromNodes, i)
		}

		if appConfig.Separator != "" {
			force := strings.Split(csvReader, appConfig.Separator)
			if (len(force) == 0) || (len(force) <= pos) {
				skip = true
			}
			if !skip {
				line = strings.Trim(force[pos], " ")
				if len(line) <= 2 {
					skip = true
				}
			}

		}

		if !skip {
			toHandle := item{&line, counter}
			for i := 0; i < len(toNodes); i++ {
				toNodes[i] <- toHandle
			}
		}

		if !skip {
			for x := 0; x < len(toNodes); x++ {
				back := <-fromNodes
				if back.score > maxScoreValue {
					maxScoreName = *(back.name)
					maxScoreValue = back.score
				}
			}
			if len(maxScoreName) > 0 {
				if outputToFile {
					stu := ResultsToWrite{Line: got, Result: maxScoreName}
					if err = pw.Write(stu); err != nil {
						log.Println("Write error", err)
					}
				} else {
					fmt.Printf("%s  %s %f\n", line, maxScoreName, maxScoreValue)
				}
			}
		}
		wg.Wait()
	}
	if outputToFile {
		if err = pw.WriteStop(); err != nil {
			log.Println("WriteStop error", err)
			return false, err
		}
	}
	return true, nil
}

func main() {

	toNodes := make([]chan item, 0)
	fromNodes := make(chan itemResult, 1)

	for i := 0; i < len(appConfig.Nodes); i++ {
		toNodes = append(toNodes, make(chan item, 1))
	}

	if _, err := ProcessFile(toNodes, fromNodes); err != nil {
		log.Fatalln("failed to run:", err.Error())
	}

}

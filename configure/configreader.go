package configure

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type NodeConfig struct {
	Name                  string     `yaml:"name"`
	Parent                string     `yaml:"parent"`
	IgnoreIfHas           []string   `yaml:"ignore_if_has"`
	MustHas               []string   `yaml:"must_has"`
	Remove                []string   `yaml:"remove"`
	ExplicitSearchCombine [][]string `yaml:"explicit_search_combine"`
	ExplicitSearch        []string   `yaml:"explicit_search"`
	FuzzySearch           string     `yaml:"fuzzy_search"`
	Threshold             float32    `yaml:"threshold"`
}

type Conf struct {
	InputFile  string
	OutputFile string
	Separator  string       `yaml:"separator"`
	Position   int          `yaml:"position"`
	Debug      bool         `yaml:"debug"`
	Nodes      []NodeConfig `yaml:"nodes"`
}

func parseCmdLineConfig(c *Conf) error {
	var (
		inputFile  string
		configFile string
		outputFile string
	)

	flag.StringVar(&configFile, "config", "", "configure yaml file")
	flag.StringVar(&inputFile, "input", "", "input file ")
	flag.StringVar(&outputFile, "output", "", "output file")

	flag.Parse()
	if inputFile == "" || configFile == "" {
		return fmt.Errorf("not enough configue")
	}

	c.InputFile = inputFile
	c.OutputFile = outputFile

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln("failed to read file " + err.Error())
		return err
	}

	err = yaml.Unmarshal([]byte(yamlFile), c)
	if err != nil {
		log.Fatalln("failed to unmarshal file " + err.Error())
		return err
	}

	return nil
}

func (c *Conf) GetConfig() (*Conf, error) {
	err := parseCmdLineConfig(c)
	return c, err
}

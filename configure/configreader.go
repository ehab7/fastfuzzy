package configure

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/ehab7/fastfuzzy/algo"
	"gopkg.in/yaml.v3"
)

type Conf struct {
	InputFile string
	Separator string   `yaml:"Separator"`
	Position  int      `yaml:"position"`
	Debug     bool     `yaml:"debug"`
	Include   []string `yaml:"include"`
	Remove    []string `yaml:"remove"`
	Reject    []string `yaml:"reject"`

	Algo algo.Algo `yaml:"algo"`
}

func parseCmdLineConfig(c *Conf) error {
	var (
		inputFile        string
		configFile       string
		separator        string
		position         int64
		debug            bool
		noSoundex        bool
		searchKeyword    string
		searchThreshold  float64
		includeKeyswords string
		removeKeywrods   string
		rejectKeywords   string
	)

	flag.StringVar(&configFile, "config", "", "configure yaml file")
	flag.StringVar(&inputFile, "input", "", "input file or stdin if file ignored")
	flag.StringVar(&searchKeyword, "search", "", "search keyword")
	flag.StringVar(&separator, "separator", "", "sparator word/char")

	flag.BoolVar(&noSoundex, "nosoundex", false, "soundex on|off if fuzzy matched")
	flag.BoolVar(&debug, "debug", false, "debug on|off")
	flag.Int64Var(&position, "position", 0, "at which positon to process the input field when seperator provided (for csv)")
	flag.Float64Var(&searchThreshold, "threshold", 0.5, "search threshold )default 0.5)")

	flag.StringVar(&includeKeyswords, "include", "", "include words")
	flag.StringVar(&removeKeywrods, "remove", "", "remove words")
	flag.StringVar(&rejectKeywords, "reject", "", "reject words")

	flag.Parse()
	c.InputFile = inputFile

	// the user provided a config yaml file
	if configFile != "" {
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

	// otherise parse config from the command line

	c.Separator = separator
	c.Position = int(position)
	c.Debug = debug
	c.Include = strings.Fields(includeKeyswords)
	c.Reject = strings.Fields(rejectKeywords)
	c.Remove = strings.Fields(removeKeywrods)
	c.Algo.Search = searchKeyword
	c.Algo.Soundex = !noSoundex
	c.Algo.Threshold = float32(searchThreshold)
	c.Algo.Debug = debug
	return nil

}

func (c *Conf) GetConfig() (*Conf, error) {

	c.Algo = algo.Algo{}

	err := parseCmdLineConfig(c)

	if err = algo.InitAlgo(&c.Algo); err != nil {
		log.Fatalln("failed to inital algo struct:" + err.Error())
	}

	if c.Debug {
		log.Println(" config struct:", fmt.Sprintf("%+v", c))
	}
	return c, nil
}

package algo

import (
	"bytes"
	"errors"
	"log"
	"strings"
)

type Algo struct {
	Search    string  `yaml:"search"`
	Threshold float32 `yaml:"threshold"`
	Soundex   bool    `yaml:"soundex"`
	Debug     bool    `yaml:"debug"`
	searchLen int
	matrix    []int
}

const codeLen = 4

var keyWordSounds = map[string]string{}

var codes = map[string]string{
	"a": "",
	"b": "1",
	"c": "2",
	"d": "3",
	"e": "",
	"f": "1",
	"g": "2",
	"h": "",
	"i": "",
	"j": "2",
	"k": "2",
	"l": "4",
	"m": "5",
	"n": "5",
	"o": "",
	"p": "1",
	"q": "2",
	"r": "6",
	"s": "2",
	"t": "3",
	"u": "",
	"v": "1",
	"w": "",
	"x": "2",
	"y": "",
	"z": "2",
}

func InitAlgo(c *Algo) error {
	if c.Search == "" {
		log.Println("no keyword to fuzzy search bail out")
		return errors.New("missing keyword")
	}

	if c.Threshold == 0.0 {
		log.Println("warning threshold less than 0.1")
		c.Threshold = 0.5
	}
	c.searchLen = len(c.Search)
	c.matrix = make([]int, c.searchLen)

	return nil
}

//  Jaro distance algorithm and allow only transposition operation
func jaroSimilarity(str1 string, config *Algo) float32 {
	// Get and store length of the strings
	str1Len := len(str1)
	str2Len := config.searchLen

	var match int
	maxStrLen := str1Len

	if str2Len > maxStrLen {
		maxStrLen = str2Len
	}

	maxDist := maxStrLen/2 - 1

	str1Table := make([]int, str1Len)

	for i := range config.matrix {
		config.matrix[i] = 0
	}

	// Check for matching characters in both strings
	for i := 0; i < str1Len; i++ {
		val1 := i - maxDist
		val2 := str2Len
		tmp := i + maxDist + 1

		if val2 > tmp {
			val2 = tmp
		}

		if val1 < 0 {
			val1 = 0
		}

		for j := val1; j < val2; j++ {
			if str1[i] == config.Search[j] && config.matrix[j] == 0 {
				str1Table[i] = 1
				config.matrix[j] = 1
				match++
				break
			}
		}
	}
	if match == 0 {
		return 0.0
	}

	var t float32
	var p int
	// Check for possible translations
	for i := 0; i < str1Len; i++ {
		if str1Table[i] == 1 {
			for config.matrix[p] == 0 {
				p++
			}
			if str1[i] != config.Search[p] {
				t++
			}
			p++
		}
	}
	t /= 2

	return (float32(match)/float32(str1Len) +
		float32(match)/float32(config.searchLen) +
		(float32(match)-t)/float32(match)) / 3.0
}

func soundex(s string) string {
	var encoded bytes.Buffer
	encoded.WriteByte(s[0])

	for i := 1; i < len(s); i++ {
		if encoded.Len() == codeLen {
			break
		}

		previous, current := string(s[i-1]), string(s[i])

		var next string
		if i+1 < len(s) {
			next = string(s[i+1])
		}

		if (current == "h" || current == "w") && (codes[previous] == codes[next]) {
			i = i + 1
			continue
		}

		if c, ok := codes[current]; ok && len(c) > 0 {
			encoded.WriteByte(c[0])
		}

		if codes[current] == codes[next] {
			i = i + 1
			continue
		}
	}

	if encoded.Len() < codeLen {
		padding := strings.Repeat("0", codeLen-encoded.Len())
		encoded.WriteString(padding)
	}

	return strings.ToUpper(encoded.String())
}

// function to apply JaroSimilarity algo the return matching status and delta from keywrod search
func Process(sentance *string, algo *Algo) (bool, float32) {

	g := strings.Fields(*sentance)
	var min = algo.Threshold
	var returnValue float32
	var found = false

	for _, item := range g {
		if len(item) <= 3 {
			continue
		}

		if item == algo.Search {
			return true, 1.01
		}

		distance := jaroSimilarity(item, algo)
		if algo.Debug {
			log.Printf("word:%s jaro calc:%f threshold:%f \n", item, distance, algo.Threshold)
		}

		if distance > returnValue {
			returnValue = distance
		}

		if distance > min {
			if algo.Soundex {

				var itemSound = soundex(item)
				kewordSound, ok := keyWordSounds[algo.Search]
				if !ok {
					kewordSound = soundex(algo.Search)
					keyWordSounds[algo.Search] = kewordSound
				}
				if algo.Debug {
					log.Printf("word:%s soundex:%s keyword:%s soundex:%s %v\n", item, itemSound, algo.Search, kewordSound, itemSound == kewordSound)
				}

				if itemSound == kewordSound {

					found = true
					min = distance
				}
			} else {
				found = true
				min = distance
			}
		}

	}

	return found, returnValue

}

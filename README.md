# Fastfuzzy

 
## This is command line/tool utility to fuzzy search in large text file or piped stream

  
- uses Jaro-Winkler distance for fuzziness calculation with optimization to allow searching large file.

- uses soundex to eliminate words picked within the search threshold but do not sound like the search keyword.

- defines separator and position for the input string help narrow the search to specific field in csv input.

- has three built in filters:
   - include:
      accepts output when fuzzy falls below threshold but above 0.5 and has anther word is the include list.
   - reject: 
      reject the output regardless the match or the fuzzy search outcome.
   - remove:
       remove certain characters from the string like '@' or '#' 

## Usage:  
```./fastfuzzy -h
Usage of fastfuzzy:
-config string    (configure yaml file)
-debug            (debug on|off)
-include string   (include words)
-input string     (input file or stdin if ignored)
-nosoundex        (soundex on|off if fuzzy matched)
-separator string (sparator char for csv)
-position int     (positon to process the input field when seperator provided for csv)
-reject string    (reject words)
-remove string    (remove words)
-search string    (search keyword)
-threshold float  (search threshold default 0.5)
```
## Examples:
Example for yaml file configuration:
```---
separator: "|"
position: 1
include:
- azzuie
reject:
- uk
- france
remove:
- '#'
- '@'
- '-'
- '/'
- '\'
debug: false
algo:
  search: beaver
  threshold: 0.85
  soundex: true
  debug: false
```

run using the config file:
```
./fastfuzzy -config myfconfig.yaml -input testfile.csv
```
anther example using the piped commandline:
  ```
cat testfile.csv | ./fastfuzzy -search "beaver" -include "lakes,ponds" -reject "forest,rivers" -remove "-,&,@" -separator "|" -position 1 --threshold 0.85 -nosoundex
```

#### Benchmarking:
Took ~110 seconds to process ~2.5G file with ~130M rows on AMD machine 2.3Ghz.

#### To-Do:
  - Adding support to rune currently it is only to English.
  - Adding support for combined words search.

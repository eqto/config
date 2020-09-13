package config

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	strRootLine = `^(?Ui)\s*\[([a-z0-9]+)\].*$`
	strLine     = `^(?Ui)\s*([a-z0-9_]+)\s*=\s*(.*)\s*((?:#|/{2,}).*|)$`
)

//Config ...
type Config struct {
	file string
	val  map[string]map[string]string
}

//Get ...
func (c *Config) Get(key string) string {
	split := strings.SplitN(key, `.`, 2)
	if c.val == nil {
		return ``
	}
	if root, ok := c.val[split[0]]; ok {
		if val, ok := root[split[1]]; ok {
			return val
		}
	}
	return ``
}

//GetInt ...
func (c *Config) GetInt(key string) int {
	str := c.Get(key)
	if str == `` {
		return 0
	}
	i, e := strconv.Atoi(str)
	if e != nil {
		return 0
	}
	return i
}

//ParseFile ...
func ParseFile(file string) (*Config, error) {
	cfg := &Config{file: file}
	f, e := os.Open(file)
	if e != nil {
		return cfg, e
	}
	return parse(f), nil
}

func parse(r io.Reader) *Config {
	scanner := bufio.NewScanner(r)
	regexLine := regexp.MustCompile(strLine)
	regexRoot := regexp.MustCompile(strRootLine)

	root := ``
	cfg := &Config{}
	for scanner.Scan() {
		strLine := scanner.Text()
		if matches := regexLine.FindStringSubmatch(strLine); len(matches) > 0 {
			key := strings.TrimSpace(matches[1])
			val := strings.TrimSpace(matches[2])
			if strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`) {
				val = val[1 : len(val)-1]
			}
			cfg.val[root][key] = val
		} else if matches := regexRoot.FindStringSubmatch(strLine); len(matches) > 0 {
			root = matches[1]
			cfg.val[root] = make(map[string]string)
		}
	}
	return cfg
}

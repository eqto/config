package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	strRootLine = `^(?Ui)\s*\[([a-z0-9]+)\].*$`
	strLine     = `^(?Ui)\s*([a-z0-9_]+)\s*=\s*(.*)(\s+(?:#|/{2,}).*|)\s*$`
)

//Config ...
type Config struct {
	file string
	val  map[string]map[string]string
}

func (c *Config) get(key string) *string {
	if c.val == nil {
		return nil
	}
	split := strings.SplitN(key, `.`, 2)
	if root, ok := c.val[split[0]]; ok {
		if val, ok := root[split[1]]; ok {
			return &val
		}
	}
	return nil
}

//Get retrieve string value by key or return empty string if not found
func (c *Config) Get(key string) string {
	if str := c.get(key); str != nil {
		return *str
	}
	return ``
}

//GetOr retrieve string value by key or return def value if not found
func (c *Config) GetOr(key, def string) string {
	if str := c.get(key); str != nil {
		return *str
	}
	return def
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

func (c *Config) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("File: %s\n", c.file))
	for rootName, rootVal := range c.val {
		sb.WriteString(fmt.Sprintf("[%s]\n", rootName))
		for key, val := range rootVal {
			sb.WriteString(fmt.Sprintf("%s = %s\n", key, val))
		}
	}
	return sb.String()
}

//ParseFile ...
func ParseFile(file string) (*Config, error) {
	cfg := &Config{file: file}
	f, e := os.Open(file)
	if e != nil {
		return cfg, e
	}
	cfg = parse(f)
	cfg.file = file
	return cfg, nil
}

func parse(r io.Reader) *Config {
	scanner := bufio.NewScanner(r)
	regexLine := regexp.MustCompile(strLine)
	regexRoot := regexp.MustCompile(strRootLine)

	root := ``
	cfg := &Config{val: make(map[string]map[string]string)}
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

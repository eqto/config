package config

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

const (
	strRootLine = `^(?Ui)\s*([-]|)\[([a-z0-9]+)\].*$`
	strLine     = `^(?Ui)\s*([a-z0-9_.]+)\s*=\s*(.*)(\s+(?:#|/{2,}).*|)\s*$`
)

var stdFile = New()

func Open(filename string) error {
	f, e := ParseFile(filename)
	if e != nil {
		return e
	}
	stdFile = f
	return nil
}

func Put(key string, value string) {
	stdFile.Put(key, value)
}

//Get retrieve string value by key or return empty string if not found
func Get(key string) string {
	return stdFile.GetOr(key, ``)
}

func GetArray(key string) []*Node {
	return stdFile.GetArray(key)
}

//GetOr retrieve string value by key or return def value if not found
func GetOr(key, def string) string {
	return stdFile.GetOr(key, def)
}

func GetInt(key string) int {
	return stdFile.GetInt(key)
}

func GetIntOr(key string, def int) int {
	return stdFile.GetIntOr(key, def)
}

func Save() error {
	return stdFile.Save()
}

func String() string {
	return stdFile.String()
}

func New() *File {
	return &File{nodes: make(map[string]*Node)}
}

func NewFile(filename string) *File {
	return &File{filename: filename, nodes: make(map[string]*Node)}
}

func ParseFile(filename string) (*File, error) {
	cfg, e := parse(filename)
	if e != nil {
		cfg = New()
		cfg.filename = filename
		return cfg, e
	}
	go cfg.watch()
	return cfg, nil
}

func parse(filename string) (*File, error) {
	f, e := os.Open(filename)
	if e != nil {
		return nil, e
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	regexLine := regexp.MustCompile(strLine)
	regexRoot := regexp.MustCompile(strRootLine)

	root := ``
	cfg := New()
	cfg.filename = filename
	for scanner.Scan() {
		strLine := scanner.Text()
		if matches := regexLine.FindStringSubmatch(strLine); len(matches) > 0 {
			key := strings.TrimSpace(matches[1])
			val := strings.TrimSpace(matches[2])
			if strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`) {
				val = val[1 : len(val)-1]
			}
			cfg.nodes[root].set(key, val)
		} else if matches := regexRoot.FindStringSubmatch(strLine); len(matches) > 0 {
			root = matches[2]
			n := new(Node)
			if matches[1] == `-` {
				if cfg.nodes[root] != nil {
					n = cfg.nodes[root]
				}
				n.arr = append(n.arr, &Node{val: make(map[string]string)})
			} else {
				n.val = make(map[string]string)
			}
			cfg.nodes[root] = n
		}
	}
	return cfg, nil
}

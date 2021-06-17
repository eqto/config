package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
)

const (
	strRootLine = `^(?Ui)\s*([-]|)\[([a-z0-9]+)\].*$`
	strLine     = `^(?Ui)\s*([a-z0-9_]+)\s*=\s*(.*)(\s+(?:#|/{2,}).*|)\s*$`
)

//Config ...
type Config struct {
	file  string
	nodes map[string]*Node
	lock  sync.RWMutex
}

//Get retrieve string value by key or return empty string if not found
func (c *Config) Get(key string) string {
	return c.GetOr(key, ``)
}

func (c *Config) GetArray(key string) []*Node {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if n, ok := c.nodes[key]; ok {
		return n.arr
	}
	return nil
}

//GetOr retrieve string value by key or return def value if not found
func (c *Config) GetOr(key, def string) string {
	if s := c.get(key); s != nil {
		return *s
	}
	return def
}

//GetInt ...
func (c *Config) GetInt(key string) int {
	return c.GetIntOr(key, 0)
}

//GetIntOr ...
func (c *Config) GetIntOr(key string, def int) int {
	if s := c.get(key); s != nil {
		i, e := strconv.Atoi(*s)
		if e != nil {
			return 0
		}
		return i
	}
	return 0
}

func (c *Config) String() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("File: %s\n", c.file))
	for nname, nval := range c.nodes {
		sb.WriteString(fmt.Sprintf("[%s]\n", nname))
		sb.WriteString(nval.String())
	}
	return sb.String()
}

func (c *Config) get(key string) *string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	split := strings.SplitN(key, `.`, 2)
	if n := c.node(split[0]); n != nil {
		if s := n.get(split[1]); s != nil {
			return s
		}
	}
	return nil
}

func (c *Config) watch() {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write)

	go func() {
		for {
			select {
			case <-w.Event:
				cfg, e := parse(c.file)
				if e == nil {
					c.lock.Lock()
					c.nodes = cfg.nodes
					c.lock.Unlock()
				}
			case err := <-w.Error:
				log.Println(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if e := w.Add(c.file); e != nil {
		log.Println(e)
	}
	if e := w.Start(1 * time.Second); e != nil {
		log.Println(e)
	}
}

func (c *Config) node(key string) *Node {
	if n, ok := c.nodes[key]; ok {
		return n
	}
	return nil
}

//ParseFile ...
func ParseFile(file string) (*Config, error) {
	cfg, e := parse(file)
	if e != nil {
		return &Config{file: file}, e
	}
	go cfg.watch()
	return cfg, nil
}

func parse(file string) (*Config, error) {
	f, e := os.Open(file)
	if e != nil {
		return nil, e
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	regexLine := regexp.MustCompile(strLine)
	regexRoot := regexp.MustCompile(strRootLine)

	root := ``
	cfg := &Config{file: file, nodes: make(map[string]*Node)}
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

package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
)

type File struct {
	filename       string
	nodes          map[string]*Node
	lock           sync.RWMutex
	changeListener func()
	watcher        *watcher.Watcher
}

// Close stops the file watcher.
func (f *File) Close() {
	if f.watcher != nil {
		f.watcher.Close()
	}
}

func (f *File) Put(key string, value string) *File {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.nodes == nil {
		f.nodes = make(map[string]*Node)
	}
	split := strings.SplitN(key, `.`, 2)
	if len(split) < 2 {
		return f
	}
	node := f.node(split[0])
	if node == nil {
		node = f.newRoot(split[0])
	}
	node.set(split[1], value)
	return f
}

// Get retrieve string value by key or return empty string if not found
func (f *File) Get(key string) string {
	return f.GetOr(key, ``)
}

func (f *File) GetArray(key string) []*Node {
	f.lock.RLock()
	defer f.lock.RUnlock()
	if n, ok := f.nodes[key]; ok {
		return n.arr
	}
	return nil
}

func (f *File) GetStringArray(key string) []string {
	f.lock.RLock()
	defer f.lock.RUnlock()

	split := strings.SplitN(key, `.`, 2)
	if len(split) < 2 {
		return nil
	}

	sectionKey, valueKey := split[0], split[1]
	if node := f.node(sectionKey); node != nil {
		if arr := node.getStringArray(valueKey); arr != nil {
			return arr
		}
		// fallback to single value
		if val := node.get(valueKey); val != nil {
			return []string{*val}
		}
	}
	return nil
}

// GetOr retrieve string value by key or return def value if not found
func (f *File) GetOr(key, def string) string {
	if s := f.get(key); s != nil {
		return *s
	}
	return def
}

func (f *File) GetInt(key string) int {
	return f.GetIntOr(key, 0)
}

func (f *File) GetIntOr(key string, def int) int {
	if s := f.get(key); s != nil {
		i, e := strconv.Atoi(*s)
		if e != nil {
			return def
		}
		return i
	}
	return def
}

func (f *File) Save() error {
	if f.filename == `` {
		return errors.New(`empty config filename`)
	}
	os.Mkdir(`configs`, 0750)
	file, e := os.OpenFile(f.filename, os.O_CREATE|os.O_WRONLY, 0750)
	if e != nil {
		return e
	}
	defer file.Close()
	if e := file.Truncate(0); e != nil {
		return e
	}
	if _, e := file.WriteString(f.String()); e != nil {
		return e
	}
	go f.watch()
	return nil
}

func (f *File) String() string {
	f.lock.RLock()
	defer f.lock.RUnlock()
	sb := strings.Builder{}
	for nname, nval := range f.nodes {
		sb.WriteString(fmt.Sprintf("[%s]\n", nname))
		sb.WriteString(nval.String())
	}
	return sb.String()
}

func (f *File) OnChanges(fun func()) {
	f.changeListener = fun
}

func (f *File) get(key string) *string {
	f.lock.RLock()
	defer f.lock.RUnlock()
	split := strings.SplitN(key, `.`, 2)
	if n := f.node(split[0]); n != nil {
		if s := n.get(split[1]); s != nil {
			return s
		}
	}
	return nil
}

func (f *File) watch() {
	f.watcher = watcher.New()
	f.watcher.SetMaxEvents(1)
	f.watcher.FilterOps(watcher.Write, watcher.Remove, watcher.Rename)

	go func() {
		for {
			select {
			case <-f.watcher.Event:
				cfg, e := parse(f.filename)
				if e == nil {
					f.lock.Lock()
					f.nodes = cfg.nodes
					f.lock.Unlock()
					if f.changeListener != nil {
						go f.changeListener()
					}
				}
				// Re-add the file to the watcher, in case it was deleted and recreated.
				f.watcher.Add(f.filename)
			case err := <-f.watcher.Error:
				log.Println(err)
			case <-f.watcher.Closed:
				return
			}
		}
	}()

	if e := f.watcher.Add(f.filename); e != nil {
		log.Println(e)
	}
	if e := f.watcher.Start(1 * time.Second); e != nil {
		log.Println(e)
	}
}

func (f *File) node(key string) *Node {
	if n, ok := f.nodes[key]; ok {
		return n
	}
	return nil
}

func (f *File) newRoot(key string) *Node {
	f.nodes[key] = &Node{val: make(map[string]nodeValue)}
	return f.node(key)
}

// Watch monitors filename for write/remove/rename events and invokes
// onChange on each — format-agnostic: unlike ParseFile it doesn't parse the
// contents, so it can watch files in any format. The returned File is just
// the lifecycle handle — Close() stops the watcher.
func Watch(filename string, onChange func()) *File {
	f := NewFile(filename)
	f.changeListener = onChange
	go f.watch()
	return f
}

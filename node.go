package config

import (
	"fmt"
	"strconv"
	"strings"
)

type Node struct {
	val map[string]string
	arr []*Node
}

func (n *Node) Get(key string) string {
	val := n.get(key)
	if val != nil {
		return *val
	}
	return ``
}

func (n *Node) GetInt(key string) int {
	val := n.get(key)
	if val != nil {
		i, _ := strconv.Atoi(*val)
		return i
	}
	return 0
}

func (n *Node) get(key string) *string {
	if n.val == nil {
		return nil
	}
	if val, ok := n.val[key]; ok {
		return &val
	}
	split := strings.SplitN(key, `.`, 2)
	if val, ok := n.val[split[0]]; ok {
		return &val
	}
	return nil
}

func (n *Node) set(key string, val string) {
	if n.val == nil {
		idx := len(n.arr) - 1
		n.arr[idx].set(key, val)
	} else {
		n.val[key] = val
	}
}

func (n *Node) String() string {
	sb := strings.Builder{}
	for key, val := range n.val {
		sb.WriteString(fmt.Sprintf("%s = %s\n", key, val))
	}
	return sb.String()

}

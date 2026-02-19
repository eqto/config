package config

import (
	"fmt"
	"strconv"
	"strings"
)

type nodeValue []string

type Node struct {
	val map[string]nodeValue
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
	if nodeVal, ok := n.val[key]; ok && len(nodeVal) > 0 {
		return &nodeVal[0]
	}
	// This part is for handling dot-separated keys.
	split := strings.SplitN(key, `.`, 2)
	if nodeVal, ok := n.val[split[0]]; ok && len(nodeVal) > 0 {
		return &nodeVal[0]
	}
	return nil
}

func (n *Node) set(key string, val string) {
	if n.val == nil {
		// This logic is for array of objects, e.g. -[database]
		if len(n.arr) > 0 {
			idx := len(n.arr) - 1
			n.arr[idx].set(key, val)
		}
		return
	}

	// Append the new value. This works even if the key doesn't exist yet.
	n.val[key] = append(n.val[key], val)
}

func (n *Node) getStringArray(key string) []string {
	if n.val != nil {
		if nodeVal, ok := n.val[key]; ok {
			return nodeVal
		}
	}
	return nil
}

func (n *Node) String() string {
	sb := strings.Builder{}
	for key, nodeVal := range n.val {
		for _, val := range nodeVal {
			sb.WriteString(fmt.Sprintf("%s = %s\n", key, val))
		}
	}
	return sb.String()
}

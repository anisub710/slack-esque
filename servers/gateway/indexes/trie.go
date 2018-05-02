package indexes

import (
	"sort"
	"strings"
	"sync"
)

//TODO: implement a trie data structure that stores
//keys of type string and values of type int64

//trieNode is a node for a trie which holds string keys mapped to the rest of the trie
// and int64set of values
type trieNode struct {
	key      rune
	children map[rune]*trieNode
	values   int64set
	prevNode *trieNode
}

//Trie is a struct for a search trie
type Trie struct {
	root   *trieNode
	mx     sync.RWMutex
	length int
}

//NewTrie constructs a new Trie.
func NewTrie() *Trie {
	return &Trie{
		root:   &trieNode{children: map[rune]*trieNode{}},
		length: 0,
	}
}

//Len returns the number of entries in the trie.
func (t *Trie) Len() int {
	return t.length
}

//Add adds a key and value to the trie.
func (t *Trie) Add(key string, value int64) {
	t.mx.Lock()
	defer t.mx.Unlock()
	currNode := t.root

	currNode = addHelper(key, currNode)
	if currNode.values == nil {
		currNode.values = int64set{}
	}
	currNode.values.add(value)
	t.length++
}

//addHelper is a helper method for the Add function
func addHelper(key string, currNode *trieNode) *trieNode {
	for _, k := range key {
		if currNode.children[k] == nil {
			newNode := &trieNode{
				children: make(map[rune]*trieNode),
			}
			currNode.children[k] = newNode
		}
		currNode.children[k].prevNode = currNode
		currNode = currNode.children[k]
		currNode.key = k
	}

	return currNode
}

//Find finds `n` values matching `prefix`. If the trie
//is entirely empty, or the prefix is empty, or n == 0,
//or the prefix is not found, this returns a nil slice.
func (t *Trie) Find(prefix string, n int) []int64 {
	t.mx.RLock()
	defer t.mx.RUnlock()
	//do checks properly
	if t.length == 0 || len(prefix) == 0 || n == 0 {
		return nil
	}
	currNode := t.root
	for _, p := range prefix {
		if currNode.children[p] == nil {
			return nil
		}
		currNode = currNode.children[p]
	}
	var result []int64
	currNode.findHelper(n, &result)
	return result

}

//findHelper is a helper method for Find which does the depth first search.
func (currNode *trieNode) findHelper(n int, result *[]int64) {
	//sort alphabetically here
	for v := range currNode.values {
		if len(*result) >= n {
			return
		}
		*result = append(*result, v)
	}
	sortedKeys := currNode.sortKeys()

	for _, k := range sortedKeys {
		if len(*result) == n {
			return
		}
		currNode.children[k].findHelper(n, result)
	}

}

func (currNode *trieNode) sortKeys() []rune {
	keys := []rune{}
	for k := range currNode.children {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

//Remove removes a key/value pair from the trie
//and trims branches with no values.
func (t *Trie) Remove(key string, value int64) {
	t.mx.Lock()
	defer t.mx.Unlock()
	currNode := t.root
	exited := false
	for _, k := range key {
		if currNode.children[k] == nil {
			exited = true
			return
		}
		currNode = currNode.children[k]

	}

	if !exited && currNode.values.has(value) {
		currNode.values.remove(value)
		t.length--
		trimBranches(currNode)
	}

}

//trimBranches removes empty
func trimBranches(currNode *trieNode) {
	if len(currNode.values) == 0 && len(currNode.children) == 0 &&
		currNode.prevNode != nil {
		key := currNode.key
		prev := currNode.prevNode
		delete(currNode.prevNode.children, key)
		trimBranches(prev)
	}
}

//convertToLowerAndSpace converts username, firstname, and lastname to lowercase
//and splits by space
func convertToLowerAndSpace(input string) []string {
	return strings.Split(strings.ToLower(input), " ")
}

//AddConvertedUsers adds each key and value pair to the trie
func (t *Trie) AddConvertedUsers(firstName string, lastName string, userName string, id int64) {
	addKeyVal(t, convertToLowerAndSpace(firstName), id)
	addKeyVal(t, convertToLowerAndSpace(lastName), id)
	addKeyVal(t, convertToLowerAndSpace(userName), id)
}

//addKeyVal adds the key and value pairs
func addKeyVal(t *Trie, result []string, id int64) {
	for _, r := range result {
		t.Add(r, id)
	}
}

//RemoveConvertedUsers removes each key and value pair to the trie
func (t *Trie) RemoveConvertedUsers(firstName string, lastName string, id int64) {
	removeKeyVal(t, convertToLowerAndSpace(firstName), id)
	removeKeyVal(t, convertToLowerAndSpace(lastName), id)
}

func removeKeyVal(t *Trie, result []string, id int64) {
	for _, r := range result {
		t.Remove(r, id)
	}

}

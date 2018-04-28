package indexes

import (
	"sync"
)

//TODO: implement a trie data structure that stores
//keys of type string and values of type int64

//trieNode is a node for a trie which holds string keys mapped to the rest of the trie
// and int64set of values
type trieNode struct {
	keys   map[rune]*trieNode
	values int64set
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
		root:   &trieNode{keys: make(map[rune]*trieNode)},
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
	//method
	for _, k := range key {
		childNode := currNode.keys[k]
		if childNode == nil {
			newNode := &trieNode{
				keys: make(map[rune]*trieNode),
			}
			currNode.keys[k] = newNode
		}
		currNode = childNode
	}
	//check if it exists before adding?
	currNode.values.add(value)
	t.length++
}

//Find finds `max` values matching `prefix`. If the trie
//is entirely empty, or the prefix is empty, or max == 0,
//or the prefix is not found, this returns a nil slice.
func (t *Trie) Find(prefix string, max int) []int64 {
	t.mx.RLock()
	defer t.mx.RUnlock()
	var found []int64
	if t.length == 0 || len(prefix) == 0 || max == 0 {
		return found
	}

	return found
	// panic("implement this function according to the comments above")
}

//Remove removes a key/value pair from the trie
//and trims branches with no values.
func (t *Trie) Remove(key string, value int64) {
	// panic("implement this function according to the comments above")
	t.mx.Lock()
	defer t.mx.Unlock()
	currNode := t.root
	for _, k := range key {
		childNode := currNode.keys[k]
		if childNode == nil {
			//do something - error?
		}
		currNode = childNode
	}
	currNode.values.remove(value)
	t.length--
	if len(currNode.values) == 0 {
		//trim branches
	}
}

package indexes

import (
	"log"
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
		root:   &trieNode{keys: map[rune]*trieNode{}},
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
		if currNode.keys[k] == nil {
			newNode := &trieNode{
				keys: make(map[rune]*trieNode),
			}
			currNode.keys[k] = newNode
		}
		currNode = currNode.keys[k]
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
		if currNode.keys[p] == nil {
			return nil
		}
		currNode = currNode.keys[p]
	}
	result := make([]int64, 0, 0)
	currNode.findHelper(n, 0, result)
	return result

}

//findHelper is a helper method for Find which does the depth first search.
func (currNode *trieNode) findHelper(n int, added int, result []int64) {
	for v := range currNode.values {
		if len(result) >= n {
			return
		}
		result = append(result, v)
	}
	for k := range currNode.keys {
		if len(result) == n {
			return
		}
		currNode.keys[k].findHelper(n, added, result)
	}

}

//Remove removes a key/value pair from the trie
//and trims branches with no values.
func (t *Trie) Remove(key string, value int64) {
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

//Dump prints out each branch in the trie
func (t *Trie) Dump() {
	currNode := t.root
	for _, k := range currNode.keys {
		log.Println(k.keys)
		// currNode.keys[k].Dump()
	}
}

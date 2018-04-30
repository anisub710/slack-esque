package indexes

import "testing"

//TODO: implement automated tests for your trie data structure
func TestAdd(t *testing.T) {
	trie := NewTrie()
	trie.Add("test", 8)
	trie.Dump()
}

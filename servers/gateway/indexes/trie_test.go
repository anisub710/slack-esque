package indexes

import (
	"reflect"
	"testing"
)

type TestKeyVal struct {
	Key string
	Val int64
}

// TODO: implement automated tests for your trie data structure
func TestAddAndFind(t *testing.T) {
	cases := []struct {
		name     string
		input    []TestKeyVal
		findQ    string
		n        int
		expected []int64
	}{
		{
			"Single Value",
			[]TestKeyVal{
				{"hello", 1},
			},
			"hello",
			2,
			[]int64{1},
		},
		{
			"Add Single Unicode",
			[]TestKeyVal{
				{"世界", 1},
			},
			"世",
			2,
			[]int64{1},
		},

		{
			"Multiple Values",
			[]TestKeyVal{
				{"hello", 1},
				{"hey", 3},
				{"tests", 8},
				{"boring", 9},
				{"hero", 10},
				{"hell", 5},
				{"世界", 1},
			},
			"he",
			4,
			[]int64{5, 1, 10, 3},
		},
		{
			"Empty Prefix",
			[]TestKeyVal{
				{"hello", 1},
				{"hey", 3},
				{"tests", 8},
				{"boring", 9},
				{"hero", 10},
				{"hell", 5},
			},
			"",
			4,
			nil,
		},
		{
			"Low limit",
			[]TestKeyVal{
				{"hello", 1},
				{"hello", 2},
				{"hello", 4},
				{"hey", 3},
				{"tests", 8},
				{"boring", 9},
				{"hero", 10},
				{"hell", 5},
			},
			"he",
			1,
			[]int64{5},
		},
		{
			"Duplicate values over limit",
			[]TestKeyVal{
				{"hello", 1},
				{"hello", 2},
				{"hello", 4},
				{"hey", 3},
				{"tests", 8},
				{"boring", 9},
				{"hero", 10},
				{"hell", 5},
				{"hell", 6},
				{"世界", 1},
			},
			"he",
			1,
			[]int64{5},
		},
		{
			"Empty Children",
			[]TestKeyVal{
				{"hello", 1},
				{"hello", 2},
				{"hello", 4},
				{"hey", 3},
				{"tests", 8},
				{"boring", 9},
				{"hero", 10},
				{"hell", 5},
				{"hell", 6},
			},
			"hellooooooooo",
			1,
			nil,
		},
	}

	for _, c := range cases {
		trie := NewTrie()
		for _, v := range c.input {
			trie.Add(v.Key, v.Val)
		}
		testResult := trie.Find(c.findQ, c.n)
		if !reflect.DeepEqual(c.expected, testResult) {
			t.Errorf("case: %s, unexpected result expected %v but got %v",
				c.name, c.expected, testResult)
		}

		if trie.Len() != len(c.input) {
			t.Errorf("case: %s, unexpected length. expected: %v, got: %v",
				c.name, c.expected, testResult)
		}
	}

}

//TODO: implement automated tests for your trie data structure
func TestRemoveAndFind(t *testing.T) {
	cases := []struct {
		name          string
		input         []TestKeyVal
		findQ         string
		toRemove      []TestKeyVal
		n             int
		expected      []int64
		expectedAfter []int64
	}{
		{
			"Remove Single Value",
			[]TestKeyVal{
				{"hello", 1},
			},
			"hello",
			[]TestKeyVal{
				{"hello", 1},
			},
			2,
			[]int64{1},
			nil,
		},
		{
			"Remove Single Unicode",
			[]TestKeyVal{
				{"世界", 1},
			},
			"世界",
			[]TestKeyVal{
				{"世界", 1},
			},
			2,
			[]int64{1},
			nil,
		},
		{
			"Remove Multiple Values",
			[]TestKeyVal{
				{"hello", 1},
				{"hey", 3},
				{"tests", 8},
				{"boring", 9},
				{"hero", 10},
				{"hell", 5},
			},
			"he",
			[]TestKeyVal{
				{"hello", 1},
				{"boring", 9},
				{"hey", 3},
			},
			4,
			[]int64{5, 1, 10, 3},
			[]int64{5, 10},
		},
		{
			"Remove Duplicate Values",
			[]TestKeyVal{
				{"hello", 1},
				{"hey", 3},
				{"tests", 8},
				{"hello", 2},
				{"boring", 9},
				{"hero", 10},
				{"hell", 5},
			},
			"he",
			[]TestKeyVal{
				{"hello", 1},
				{"hey", 3},
			},
			5,
			[]int64{5, 1, 2, 10, 3},
			[]int64{5, 2, 10},
		},
		{
			"Remove Value That Don't Exist",
			[]TestKeyVal{
				{"hello", 1},
				{"hey", 3},
				{"tests", 8},
				{"hello", 2},
				{"boring", 9},
				{"hero", 10},
				{"hell", 5},
			},
			"he",
			[]TestKeyVal{
				{"hellooo", 1},
			},
			5,
			[]int64{5, 1, 2, 10, 3},
			[]int64{5, 1, 2, 10, 3},
		},
	}

	for _, c := range cases {
		trie := NewTrie()
		for _, v := range c.input {
			trie.Add(v.Key, v.Val)
		}
		testResult := trie.Find(c.findQ, c.n)
		if !reflect.DeepEqual(c.expected, testResult) {
			t.Errorf("case: %s, unexpected result expected %v, got %v", c.name, c.expected, testResult)
		}
		for _, v := range c.toRemove {
			trie.Remove(v.Key, v.Val)
		}
		testResultAfter := trie.Find(c.findQ, c.n)
		if !reflect.DeepEqual(c.expectedAfter, testResultAfter) {
			t.Errorf("case: %s, unexpected result after: expected %v, got %v", c.name, c.expectedAfter, testResultAfter)
		}
	}

}

func TestAddConvertedUsers(t *testing.T) {
	cases := []struct {
		name      string
		firstName string
		lastName  string
		userName  string
		id        int64
		findQ     string
		expected  []int64
	}{
		{
			"Add normal",
			"Competent",
			"Gopher",
			"test1234",
			1,
			"c",
			[]int64{1},
		},
		{
			"Add unicode",
			"世界",
			"Gopher",
			"test1234",
			1,
			"世",
			[]int64{1},
		},
		{
			"Add space",
			"multiple first",
			"Gopher",
			"test1234",
			1,
			"m",
			[]int64{1},
		},
	}

	for _, c := range cases {
		trie := NewTrie()
		trie.AddConvertedUsers(c.firstName, c.lastName, c.userName, c.id)
		testResult := trie.Find(c.findQ, 1)
		if !reflect.DeepEqual(c.expected, testResult) {
			t.Errorf("case: %s, unexpected result expected %v, got %v", c.name, c.expected, testResult)
		}
	}

}

func TestRemoveConvertedUsers(t *testing.T) {
	cases := []struct {
		name          string
		firstName     string
		lastName      string
		userName      string
		input         []TestKeyVal
		id            int64
		findQ         string
		expected      []int64
		expectedAfter []int64
	}{
		{
			"Remove normal",
			"Competent",
			"Gopher",
			"test1234",
			[]TestKeyVal{
				{"competent", 1},
				{"gopher", 1},
			},
			1,
			"c",
			[]int64{1},
			nil,
		},

		{
			"Remove unicode",
			"世界",
			"Gopher",
			"test1234",
			[]TestKeyVal{
				{"世界", 1},
				{"gopher", 1},
			},
			1,
			"世",
			[]int64{1},
			nil,
		},
		{
			"Remove unicode",
			"multiple first",
			"Gopher",
			"test1234",
			[]TestKeyVal{
				{"multiple", 1},
				{"first", 1},
				{"gopher", 1},
			},
			1,
			"m",
			[]int64{1},
			nil,
		},
	}

	for _, c := range cases {
		trie := NewTrie()
		for _, v := range c.input {
			trie.Add(v.Key, v.Val)
		}
		testResult := trie.Find(c.findQ, 1)
		if !reflect.DeepEqual(c.expected, testResult) {
			t.Errorf("case: %s, unexpected result expected %v, got %v", c.name, c.expected, testResult)
		}
		trie.RemoveConvertedUsers(c.firstName, c.lastName, c.id)
		testResultAfter := trie.Find(c.findQ, 1)
		if !reflect.DeepEqual(c.expectedAfter, testResultAfter) {
			t.Errorf("case: %s, unexpected result expected %v, got %v", c.name, c.expected, testResult)
		}

	}

}

package model

type ClientSet map[*Client]struct{}

func NewClientSet() ClientSet {
	return make(map[*Client]struct{})
}

func (set ClientSet) Add(c *Client) {
	set[c] = struct{}{}
}

func (set ClientSet) IsMember(c *Client) bool {
	if _, ok := set[c]; ok {
		return true
	}
	return false
}

func (set ClientSet) Remove(c *Client) {
	if _, ok := set[c]; !ok {
		return
	}
	delete(set, c)
}

func (set ClientSet) Count() int {
	return len(set)
}

func (set ClientSet) Clone() ClientSet {
	n := make(map[*Client]struct{})
	for k, v := range set {
		n[k] = v
	}
	return n
}

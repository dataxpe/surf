package jar

import (
	"sync"

	"github.com/Diggernaut/goquery"
)

type AsyncStore struct {
	sync.Mutex
	doms map[string]*goquery.Document
}

func NewAsyncStore() *AsyncStore {
	return &AsyncStore{doms: make(map[string]*goquery.Document)}
}
func (a *AsyncStore) Set(name string, val *goquery.Document) {
	a.Lock()
	defer a.Unlock()
	a.doms[name] = val
}
func (a *AsyncStore) Get(name string) *goquery.Document {
	a.Lock()
	defer a.Unlock()
	if v, ok := a.doms[name]; ok {
		return v
	}
	return &goquery.Document{}
}

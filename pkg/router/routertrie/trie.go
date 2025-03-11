package routertrie

import (
	"context"
	"net/http"
	"strings"
)

var (
	paramMapId = "{$param$}"
)

type HttpMethod string

type ChildsNode map[string]*Node
type HandlersMap map[HttpMethod]http.HandlerFunc

type Node struct {
	Segment   string
	IsParam   bool
	ParamName string
	Childs    ChildsNode
	Handlers  HandlersMap
}

type Trie struct {
	root *Node
}

func NewTrie() *Trie {
	return &Trie{
		root: &Node{
			Segment:   "",
			IsParam:   false,
			ParamName: "",
			Childs:    map[string]*Node{},
			Handlers:  map[HttpMethod]http.HandlerFunc{},
		},
	}
}

func (t *Trie) Put(method, path string, handler http.HandlerFunc) {
	if path == "/" {
		t.root.Segment = path
		t.root.Handlers[HttpMethod(method)] = handler
		return
	}

	segments := strings.Split(strings.Trim(path, "/"), "/")

	curNode := t.root
	for _, seg := range segments {
		pName, isPar := t.isParamSegment(seg)

		segMapId := seg
		if isPar {
			segMapId = paramMapId
		}

		n, exist := curNode.Childs[segMapId]

		if exist {
			curNode = n
			continue
		}

		newNode := &Node{
			Segment:   seg,
			Childs:    map[string]*Node{},
			IsParam:   isPar,
			ParamName: pName,
		}

		curNode.Childs[segMapId] = newNode
		curNode = newNode
	}

	if curNode.Handlers == nil {
		curNode.Handlers = HandlersMap{}
	}

	curNode.Handlers[HttpMethod(method)] = handler
}

func (t *Trie) isParamSegment(seg string) (name string, isParam bool) {
	if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
		return string(seg[1 : len(seg)-1]), true
	}

	return "", false
}

type CtxParamKey string

func (t *Trie) FindHandler(r *http.Request) (http.HandlerFunc, *http.Request) {
	ctx := r.Context()

	path := r.URL.Path
	method := r.Method
	if path == "/" {
		return t.root.Handlers[HttpMethod(method)], r
	}

	segments := strings.Split(strings.Trim(path, "/"), "/")
	curNode := t.root
	for _, seg := range segments {
		node, exist := curNode.Childs[seg]

		if !exist {
			node = curNode.Childs[paramMapId]
			if node == nil {
				return nil, r
			}

			pKey := node.ParamName
			ctx = context.WithValue(ctx, CtxParamKey(pKey), seg)
		}

		curNode = node
	}

	return curNode.Handlers[HttpMethod(method)], r.WithContext(ctx)
}

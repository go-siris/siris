package node

import (
	//"fmt"
	"sort"
	"strings"

	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/core/errors"
)

// Nodes a conversion type for []*node.
type Nodes []*node

type node struct {
	s                 string
	wildcardParamName string   // name of the wildcard parameter, only one per whole Node is allowed
	paramNames        []string // only-names
	childrenNodes     Nodes
	handlers          context.Handlers
	root              bool
	rootWildcard      bool // if it's a wildcard {path} type on root, it should allow everything but it is not conflicts with
	// any other static or dynamic or wildcard paths if exists on other nodes.
}

// ErrDublicate returnned from `Add` when two or more routes have the same registered path.
var ErrDublicate = errors.New("two or more routes have the same registered path")

// Add adds a node to the tree, returns an ErrDublicate error on failure.
func (nodes *Nodes) Add(path string, handlers context.Handlers) error {
	//fmt.Println("[Add] adding path: " + path)
	// resolve params and if that node should be added as root
	var params []string
	var paramStart, paramEnd int
	for {
		paramStart = strings.IndexByte(path[paramEnd:], ':')
		if paramStart == -1 {
			break
		}
		paramStart += paramEnd
		paramStart++
		paramEnd = strings.IndexByte(path[paramStart:], '/')

		if paramEnd == -1 {
			params = append(params, path[paramStart:])
			path = path[:paramStart]
			break
		}
		paramEnd += paramStart
		params = append(params, path[paramStart:paramEnd])
		path = path[:paramStart] + path[paramEnd:]
		paramEnd -= paramEnd - paramStart
	}

	var p []int
	for i := 0; i < len(path); i++ {
		idx := strings.IndexByte(path[i:], ':')
		if idx == -1 {
			break
		}
		p = append(p, idx+i)
		i = idx + i
	}

	for _, idx := range p {
		//fmt.Print("-2 nodes.Add: path: " + path + " params len: ")
		//fmt.Println(len(params))
		if err := nodes.add(path[:idx], nil, nil, true); err != nil {
			return err
		}
		//fmt.Print("-1 nodes.Add: path: " + path + " params len: ")
		//fmt.Println(len(params))
		if nidx := idx + 1; len(path) > nidx {
			if err := nodes.add(path[:nidx], nil, nil, true); err != nil {
				return err
			}
		}
	}

	//fmt.Print("nodes.Add: path: " + path + " params len: ")
	//fmt.Println(len(params))
	if err := nodes.add(path, params, handlers, true); err != nil {
		return err
	}

	// prioritize by static path remember, they were already sorted by subdomains too.
	nodes.prioritize()
	return nil
}

func (nodes *Nodes) add(path string, paramNames []string, handlers context.Handlers, root bool) (err error) {
	//fmt.Println("-----------")
	//fmt.Printf("[add] adding path: %#v\n", path)
	// wraia etsi doulevei ara
	// na to kanw na exei to node to diko tou wildcard parameter name
	// kai sto telos na pernei auto, me vasi to *paramname
	// alla edw mesa 9a ginete register vasi tou last /

	// set the wildcard param name to the root and its children.
	wildcardIdx := strings.IndexByte(path, '*')
	wildcardParamName := ""
	if wildcardIdx > 0 {
		wildcardParamName = path[wildcardIdx+1:]

		path = path[0:wildcardIdx-1] + "/" // replace *paramName with single slash
		//fmt.Printf("[add] change path: %#v\n", path)
		//fmt.Printf("[add] adding wildcardParamName: %#v\n", wildcardParamName)
		//fmt.Printf("[add] adding paramNames: %#v\n", paramNames)
	}

	if wildcardIdx > 0 && wildcardParamName != "" && root {
		n := &node{
			s:                 path,
			wildcardParamName: wildcardParamName,
			paramNames:        paramNames,
			handlers:          handlers,
			root:              root,
		}
		// if root wildcard, then add it as it's and return
		if path == "/" {
			path += "/" // if root wildcard, then do it like "//" instead of simple "/"
			n.s = path
			n.rootWildcard = true

		} else {
			n.rootWildcard = false
		}
		*nodes = append(*nodes, n)
		return
	}

loop:
	for _, n := range *nodes {
		if n.rootWildcard {
			continue
		}

		minlen := len(n.s)
		if len(path) < minlen {
			minlen = len(path)
		}

		for i := 0; i < minlen; i++ {
			if n.s[i] == path[i] {
				continue
			}
			if i == 0 {
				continue loop
			}

			*n = node{
				s: n.s[:i],
				childrenNodes: Nodes{
					{
						s:                 n.s[i:],
						wildcardParamName: n.wildcardParamName, // wildcardParamName
						paramNames:        n.paramNames,
						childrenNodes:     n.childrenNodes,
						handlers:          n.handlers,
					},
					{
						s:                 path[i:],
						wildcardParamName: wildcardParamName,
						paramNames:        paramNames,
						handlers:          handlers,
					},
				},
				root: n.root,
			}

			// //fmt.Printf("%#v\n", n)
			//fmt.Println("2. change n and return  " + n.s[:i] + " and " + path[i:])
			return
		}

		if len(path) < len(n.s) {
			//fmt.Println("3. change n and return | n.s[:len(path)] = " + n.s[:len(path)-1] + " and child: " + n.s[len(path)-1:])

			*n = node{
				s:                 n.s[:len(path)],
				wildcardParamName: wildcardParamName,
				paramNames:        paramNames,
				childrenNodes: Nodes{
					{
						s:                 n.s[len(path):],
						wildcardParamName: n.wildcardParamName, // wildcardParamName
						paramNames:        n.paramNames,
						childrenNodes:     n.childrenNodes,
						handlers:          n.handlers,
					},
				},
				handlers: handlers,
				root:     n.root,
			}

			return
		}

		if len(path) > len(n.s) {
			if n.wildcardParamName != "" {
				n := &node{
					s:                 path,
					wildcardParamName: wildcardParamName,
					paramNames:        paramNames,
					handlers:          handlers,
					root:              root,
				}
				//fmt.Println("3.5. nodes.Add path: " + n.s)
				*nodes = append(*nodes, n)
				return
			}
			//fmt.Println("4. nodes.Add path: " + path[len(n.s):])
			err = n.childrenNodes.add(path[len(n.s):], paramNames, handlers, false)
			return err
		}

		if len(handlers) == 0 { // missing handlers
			return nil
		}

		if len(n.handlers) > 0 { // n.handlers already setted
			return ErrDublicate
		}
		n.paramNames = paramNames
		n.handlers = handlers

		return
	}

	n := &node{
		s:                 path,
		wildcardParamName: wildcardParamName,
		paramNames:        paramNames,
		handlers:          handlers,
		root:              root,
	}
	//fmt.Println("5. nodes.Add path: " + n.s)
	*nodes = append(*nodes, n)
	return
}

// Find resolves the path, fills its params
// and returns the registered to the resolved node's handlers.
func (nodes Nodes) Find(path string, params *context.RequestParams) context.Handlers {
	n, paramValues := nodes.findChild(path, nil)
	if n != nil {
		//	map the params,
		// n.params are the param names
		if len(paramValues) > 0 {
			//fmt.Println("-----------")
			//fmt.Print("param values returned len: ")
			//fmt.Println(len(paramValues))
			//fmt.Println("first value is: " + paramValues[0])
			//fmt.Print("n.paramNames len: ")
			//fmt.Println(len(n.paramNames))
			//fmt.Print("n.wildcardParamName len: ")
			//fmt.Println(len(n.wildcardParamName))
			for i, name := range n.paramNames {
				//fmt.Println("setting param name: " + name + " = " + paramValues[i])
				params.Set(name, paramValues[i])
			}
			// last is the wildcard,
			// if paramValues are exceed from the registered param names.
			// Note that n.wildcardParamName can be not empty but that doesn't meaning
			// that it contains a wildcard path, so the check is required.
			if len(paramValues) > len(n.paramNames) {
				//fmt.Printf("len(paramValues) %d > %d len(n.paramNames)", len(paramValues), len(n.paramNames))
				lastWildcardVal := paramValues[len(paramValues)-1]
				//if n.wildcardParamName == "" {
				//	n.wildcardParamName = "file"
				//}
				//fmt.Println("setting wildcard param name: " + n.wildcardParamName + " = " + lastWildcardVal)
				params.Set(n.wildcardParamName, lastWildcardVal)
			}
		}
		return n.handlers
	}

	return nil
}

// Exists returns true if a node with that "path" exists,
// otherise false.
//
// We don't care about parameters here.
func (nodes Nodes) Exists(path string) bool {
	n, _ := nodes.findChild(path, nil)
	return n != nil && len(n.handlers) > 0
}

func (nodes Nodes) findChild(path string, params []string) (*node, []string) {
	//fmt.Println("request path: " + path)
	for _, n := range nodes {
		if n.s == ":" {
			paramEnd := strings.IndexByte(path, '/')
			if paramEnd == -1 {
				if len(n.handlers) == 0 {
					return nil, nil
				}
				return n, append(params, path)
			}
			return n.childrenNodes.findChild(path[paramEnd:], append(params, path[:paramEnd]))
		}
		//fmt.Printf("n: %#v\n", n)
		//fmt.Println("n.s: " + n.s)
		//fmt.Print("n.childrenNodes len: ")
		//fmt.Println(len(n.childrenNodes))
		//fmt.Print("n.root: ")
		//fmt.Println(n.root)

		// by runtime check of:,
		// if n.s == "//" && n.root && n.wildcardParamName != "" {
		// but this will slow down, so we have a static field on the node itself:
		if n.rootWildcard {
			//fmt.Println("return from n.rootWildcard")
			// single root wildcard
			if len(path) < 2 {
				// do not remove that, it seems useless but it's not,
				// we had an error while production, this fixes that.
				path = "/" + path
			}
			return n, append(params, path[1:])
		}

		if !strings.HasPrefix(path, n.s) {
			// //fmt.Printf("---here root: %v, n.s: "+n.s+" and path: "+path+" is dynamic: %v , wildcardParamName: %s, children len: %v \n", n.root, n.isDynamic(), n.wildcardParamName, len(n.childrenNodes))
			continue
		}

		if len(path) == len(n.s) {
			if len(n.handlers) == 0 {
				return nil, nil
			}
			return n, params
		}

		child, childParamNames := n.childrenNodes.findChild(path[len(n.s):], params)
		//fmt.Print("childParamNames len: ")
		//fmt.Println(len(childParamNames))

		if len(childParamNames) > 0 {
			//fmt.Println("childParamsNames[0] = " + childParamNames[0])
		}

		if child == nil || len(child.handlers) == 0 {
			if n.s[len(n.s)-1] == '/' && !(n.root && (n.s == "/" || len(n.childrenNodes) > 0)) {
				if len(n.handlers) == 0 {
					return nil, nil
				}

				//fmt.Println("if child == nil.... | n.s = " + n.s)
				//fmt.Print("n.paramNames len: ")
				//fmt.Println(n.paramNames)
				//fmt.Print("n.wildcardParamName is: ")
				//fmt.Println(n.wildcardParamName)
				//fmt.Print("return n, append(params, path[len(n.s) | params: ")
				//fmt.Println(path[len(n.s):])
				return n, append(params, path[len(n.s):])
			}

			continue
		}

		return child, childParamNames
	}
	return nil, nil
}

// childLen returns all the children's and their children's length.
func (n *node) childLen() (i int) {
	for _, n := range n.childrenNodes {
		i++
		i += n.childLen()
	}
	return
}

func (n *node) isDynamic() bool {
	return n.s == ":" || n.wildcardParamName != "" || n.rootWildcard
}

// prioritize sets the static paths first.
func (nodes Nodes) prioritize() {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].isDynamic() {
			return false
		}
		if nodes[j].isDynamic() {
			return true
		}

		return nodes[i].childLen() > nodes[j].childLen()
	})

	for _, n := range nodes {
		n.childrenNodes.prioritize()
	}
}

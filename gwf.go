package gwf

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	gwfLog "github.com/fz18/gwf/log"
)

type HandlerFunc func(c *Context)

type MiddleWare func(HandlerFunc) HandlerFunc

type routerGroup struct {
	name        string
	handlerMap  map[string]map[string]HandlerFunc
	tree        *treeNode
	middlewares []MiddleWare
}

func (r *routerGroup) methodHandle(ctx *Context, h HandlerFunc) {
	for _, mid := range r.middlewares {
		h = mid(h)
	}
	h(ctx)
}

func (r *routerGroup) Use(m ...MiddleWare) {
	r.middlewares = append(r.middlewares, m...)
}

func (r *routerGroup) handle(name, method string, handler HandlerFunc) {
	_, ok := r.handlerMap[name]
	if !ok {
		r.handlerMap[name] = map[string]HandlerFunc{}
	}
	_, ok = r.handlerMap[name][method]
	if ok {
		panic("route dumplicate")
	}
	r.handlerMap[name][method] = handler
	r.tree.Put(name)
}

func (r *routerGroup) Any(name string, handler HandlerFunc) {
	r.handle(name, "Any", handler)
}
func (r *routerGroup) Get(name string, handler HandlerFunc) {
	r.handle(name, http.MethodGet, handler)
}
func (r *routerGroup) Post(name string, handler HandlerFunc) {
	r.handle(name, http.MethodPost, handler)
}

type router struct {
	groups []*routerGroup
	engine *Engine
}

func (r *router) Group(name string) *routerGroup {
	group := &routerGroup{
		name:       name,
		handlerMap: make(map[string]map[string]HandlerFunc),
		tree:       &treeNode{name: "/", children: make([]*treeNode, 0)},
	}
	r.groups = append(r.groups, group)
	group.Use(r.engine.middlewares...)
	return group
}

type Engine struct {
	router
	pool        sync.Pool
	middlewares []MiddleWare
	Logger      *gwfLog.Logger
}

func New() *Engine {
	engine := &Engine{router: router{}}
	engine.pool.New = engine.allocateContext
	return engine
}

func Default() *Engine {
	engine := New()
	engine.Logger = gwfLog.Default()
	engine.router.engine = engine
	return engine
}

func (e *Engine) Use(midwares ...MiddleWare) {
	e.middlewares = append(e.middlewares, midwares...)
}

func (e *Engine) allocateContext() any {
	return &Context{engine: e}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := e.pool.Get().(*Context)
	ctx.W = w
	ctx.R = r
	ctx.Logger = e.Logger
	e.httpRequestHandle(ctx)
	e.pool.Put(ctx)
}

func (e *Engine) httpRequestHandle(ctx *Context) {
	w := ctx.W
	r := ctx.R
	method := r.Method
	for _, group := range e.router.groups {
		routerName := SubStringLast(r.URL.Path, "/"+group.name)
		node := group.tree.Get(routerName)
		if node != nil && node.isEnd {
			// any match
			handler, ok := group.handlerMap[node.path]["Any"]
			if !ok {
				handler, ok = group.handlerMap[node.path][method]
			}
			if ok {
				group.methodHandle(ctx, handler)
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "%s %s not allowed", r.RequestURI, r.Method)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s not found", r.RequestURI)

}

func (e *Engine) Run(addr string) {

	http.Handle("/", e)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

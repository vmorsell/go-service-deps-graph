package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Graph struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`

	nodes map[string]struct{}
	links map[string]string
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]struct{}),
		links: make(map[string]string),
	}
}

func (g *Graph) AddNode(node Node) {
	if _, ok := g.nodes[node.ID]; ok {
		return
	}
	g.Nodes = append(g.Nodes, node)
	g.nodes[node.ID] = struct{}{}
}

func (g *Graph) AddLink(link Link) {
	if target, ok := g.links[link.Source]; ok {
		if target == link.Target {
			return
		}
	}
	g.Links = append(g.Links, link)
	g.links[link.Source] = link.Target
}

type Node struct {
	ID string `json:"id"`
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: go run main.go /path/to/repo/root")
	}

	path := os.Args[1]

	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalf("read dir: %v", err)
	}

	graph := NewGraph()

	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		if !strings.HasPrefix(d.Name(), "go-service") {
			continue
		}

		if d.Name() == "go-service" {
			continue
		}

		f, err := os.ReadFile(fmt.Sprintf("%s/%s/go.mod", path, d.Name()))
		if err != nil {
			log.Fatalf("read modfile: %v", err)
		}

		graph.AddNode(Node{
			ID: d.Name(),
		})
		log.Printf("discovering service %s", d.Name())

		re := regexp.MustCompile(`github\.com\/northvolt\/(go\-service\-[a-z\-]*) v[0-9\.]*\n`)
		lines := re.FindAllSubmatch(f, -1)
		for _, l := range lines {
			if len(l) != 2 {
				log.Fatalf("unexpected size %d of regexp match for line %s", len(l), l)
			}
			dep := string(l[1])
			graph.AddNode(Node{
				ID: dep,
			})
			graph.AddLink(Link{
				Source: d.Name(),
				Target: dep,
				Value:  1,
			})
			log.Printf(" - %s", dep)
		}
	}

	jsonGraph, err := json.Marshal(graph)
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}

	file, err := os.Create("graph.json")
	if err != nil {
		log.Fatalf("create: %v", err)
	}

	_, err = file.Write(jsonGraph)
	if err != nil {
		log.Fatalf("write: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	const port = 5555
	log.Printf("http://localhost:%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}

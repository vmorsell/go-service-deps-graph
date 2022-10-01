package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

// Nodes represents services.
type Node struct {
	Name  string
	Edges []Edge
}

// Edges represents dependencies between services.
type Edge struct {
	From      *Node
	To        *Node
	Direction Direction

	// Todo: calculate weight of edges
}

type Direction int

const (
	In Direction = iota
	Out
)

func main() {
	nodes := make(map[string]*Node)

	if len(os.Args) < 2 {
		log.Fatalf("usage: go run main.go /path/to/repo/root")
	}

	path := os.Args[1]

	repoDirs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalf("read dir: %v", err)
	}

	for _, repo := range repoDirs {
		if !repo.IsDir() {
			continue
		}

		if !strings.HasPrefix(repo.Name(), "go-service") {
			continue
		}

		if repo.Name() == "go-service" || repo.Name() == "go-service-attachment" {
			continue
		}

		log.Printf("reading %s", repo.Name())

		to, ok := nodes[repo.Name()]
		if !ok {
			to = &Node{
				Name: repo.Name(),
			}
			nodes[to.Name] = to
		}

		f, err := os.ReadFile(fmt.Sprintf("%s/%s/go.mod", path, repo.Name()))
		if err != nil {
			log.Fatalf("read modfile: %v", err)
		}

		re := regexp.MustCompile(`github\.com\/northvolt\/(go\-service\-[a-z\-]*) v[0-9\.]*\n`)
		lines := re.FindAllSubmatch(f, -1)
		for _, l := range lines {
			if len(l) != 2 {
				log.Fatalf("unexpected size %d of regexp match for line %s", len(l), l)
			}
			fromName := string(l[1])

			from, ok := nodes[fromName]
			if !ok {
				from = &Node{
					Name: fromName,
				}
				nodes[from.Name] = from
			}

			nodes[from.Name].Edges = append(nodes[from.Name].Edges, Edge{
				From:      from,
				To:        to,
				Direction: Out,
			})
			nodes[to.Name].Edges = append(nodes[to.Name].Edges, Edge{
				From:      from,
				To:        to,
				Direction: In,
			})
		}
	}

	for _, n := range nodes {
		var in []Edge
		var out []Edge

		for _, e := range n.Edges {
			switch e.Direction {
			case In:
				in = append(in, e)
			case Out:
				out = append(out, e)
			}
		}

		log.Printf("%s (%d deps / used in %d)", n.Name, len(in), len(out))

		for _, e := range in {
			log.Printf(" - %s <- %s", e.To.Name, e.From.Name)
		}
		for _, e := range out {
			log.Printf(" - %s -> %s", e.From.Name, e.To.Name)
		}
	}
}

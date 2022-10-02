package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
)

// Service struct.
type Service struct {
	Name string
	Deps []string
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: go run main.go /path/to/repo/root")
	}

	path := os.Args[1]

	repoDirs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalf("read dir: %v", err)
	}

	var services []Service

	for _, repo := range repoDirs {
		if !repo.IsDir() {
			continue
		}

		if !strings.HasPrefix(repo.Name(), "go-service") {
			continue
		}

		if repo.Name() == "go-service" {
			continue
		}

		f, err := os.ReadFile(fmt.Sprintf("%s/%s/go.mod", path, repo.Name()))
		if err != nil {
			log.Fatalf("read modfile: %v", err)
		}

		svc := Service{
			Name: repo.Name(),
		}

		re := regexp.MustCompile(`github\.com\/northvolt\/(go\-service\-[a-z\-]*) v[0-9\.]*\n`)
		lines := re.FindAllSubmatch(f, -1)
		for _, l := range lines {
			if len(l) != 2 {
				log.Fatalf("unexpected size %d of regexp match for line %s", len(l), l)
			}
			dep := string(l[1])
			svc.Deps = append(svc.Deps, dep)
		}

		log.Printf("discovered service: %s (%d deps)", svc.Name, len(svc.Deps))
		for i, d := range svc.Deps {
			log.Printf("  %d: %s", i, d)
		}
		services = append(services, svc)
	}

	dataset := d3Dataset(services)
	if err := writeGraph(dataset); err != nil {
		log.Fatalf("write graph: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "graph.html")
	})

	const port = 5555
	log.Printf("http://localhost:%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}

type D3Dataset struct {
	Nodes []string
	Edges [][]int
}

func d3Dataset(services []Service) D3Dataset {
	nodes := make([]string, 0, len(services))
	edges := make([][]int, len(services))
	for i, s := range services {
		nodes = append(nodes, s.Name)
		edges[i] = make([]int, len(services))

		for j, ss := range services {
			if i == j {
				continue
			}

			for _, d := range s.Deps {
				if d == ss.Name {
					edges[i][j] = 1
				}
			}
		}
	}
	return D3Dataset{
		Nodes: nodes,
		Edges: edges,
	}
}

func writeGraph(dataset D3Dataset) error {
	tpl, err := ioutil.ReadFile("template.html.gotpl")
	if err != nil {
		return fmt.Errorf("read template file: %w", err)
	}

	t := template.New("tpl")

	t.Funcs(map[string]interface{}{
		"jsStringArray": func(in []string) string {
			return fmt.Sprintf("\"%s\"", strings.Join(in, "\",\""))
		},
		"jsIntMatrix": func(in [][]int) string {
			b := bytes.NewBuffer(nil)
			for i, r := range in {
				fmt.Fprint(b, "[")
				for j, c := range r {
					fmt.Fprint(b, c)
					if j < len(r)-1 {
						fmt.Fprint(b, ",")
					}
				}
				fmt.Fprint(b, "]")
				if i < len(in)-1 {
					fmt.Fprint(b, ",")
				}
			}
			return b.String()
		},
	})

	t, err = t.Parse(string(tpl))
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	f, err := os.Create("graph.html")
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	err = t.Execute(f, dataset)
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}

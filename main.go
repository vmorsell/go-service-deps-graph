package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
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

		if repo.Name() == "go-service" || repo.Name() == "go-service-attachment" {
			continue
		}

		log.Printf("reading %s", repo.Name())

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

		services = append(services, svc)
	}

	for _, s := range services {
		log.Printf("%s (%d deps)", s.Name, len(s.Deps))
		for _, d := range s.Deps {
			log.Printf(" - %s -> %s", s.Name, d)
		}
	}
}

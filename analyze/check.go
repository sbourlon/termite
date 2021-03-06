package analyze

import (
	"fmt"
	"log"
)

type targetSet map[*Target]struct{}

var yes = struct{}{}

type undeclaredDep struct {
	Target *String
	Read   *String
}

func (u *undeclaredDep) HTML(g *Graph) string {
	return fmt.Sprintf("undeclared dependency: target %s reads %s",
		g.targetURL(u.Target), u.Read)
}

type unusedDep struct {
	Target *String
	Dep    *String
}

func (u *unusedDep) HTML(g *Graph) string {
	return fmt.Sprintf("unused dependency: target %s, dep %s",
		g.targetURL(u.Target), u.Dep)
}

type checkTargetResult struct {
	errors []Error
	edges  map[edge]struct{}
}

func (g *Graph) checkTarget(target *Target) checkTargetResult {
	realDeps := map[*Target]*String{}
	for r := range target.Reads {
		t := g.TargetByWrite[r]
		if t != nil {
			realDeps[t] = r
		}
	}
	result := checkTargetResult{
		edges: make(map[edge]struct{}),
	}

	usedDeps := targetSet{}
	for dep, name := range realDeps {
		if _, ok := target.Deps[name]; ok {
			result.edges[edge{target, dep}] = yes
			usedDeps[dep] = yes
		}
		if _, ok := target.Deps[dep.Name]; ok {
			usedDeps[dep] = yes
		}
	}
	for d := range usedDeps {
		delete(realDeps, d)
	}

	for d := range realDeps {
		path := g.findPath([]*Target{target}, d)
		if path != nil {
			for i := 1; i < len(path); i++ {
				result.edges[edge{path[i-1], path[i]}] = yes
			}
		} else {
			e := &undeclaredDep{target.Name, realDeps[d]}
			result.errors = append(result.errors, e)
			target.Errors = append(target.Errors, e)
		}
	}

	return result
}

// finds a path reaching needle from the partial path given. Returns
// the complete path or nil.
func (g *Graph) findPath(partial []*Target, needle *Target) []*Target {
	target := partial[len(partial)-1]
	if target == needle {
		return partial
	}

nextDep:
	for d := range target.Deps {
		dep := g.TargetByName[d]
		if dep == nil {
			continue
		}
		for _, done := range partial {
			if done == dep {
				log.Println("cyclic dep", target.Name, dep.Name)
				continue nextDep
			}
		}

		if try := g.findPath(append(partial, dep), needle); try != nil {
			return try
		}
	}
	return nil
}

func (g *Graph) checkTargets() {
	log.Println("checking targets")
	results := make(chan checkTargetResult, 100)
	for _, target := range g.TargetByName {
		go func(t *Target) {
			results <- g.checkTarget(t)
		}(target)
	}

	for _ = range g.TargetByName {
		r := <-results
		for e := range r.edges {
			g.UsedEdges[e] = yes
		}
		g.Errors = append(g.Errors, r.errors...)
	}

	for _, target := range g.TargetByName {
		g.checkUnusedDeps(target)
	}
	log.Println("done checking targets")
}

func (g *Graph) checkUnusedDeps(target *Target) {
	for dep := range target.Deps {
		depTarget := g.TargetByName[dep]
		if _, ok := g.UsedEdges[edge{target, depTarget}]; !ok {
			target.Errors = append(target.Errors, &unusedDep{
				target.Name, dep})
		}
	}
}

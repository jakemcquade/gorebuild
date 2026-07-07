package main

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	buildMu sync.Mutex
	builds  = map[string]*build{}
)

type build struct {
	cancel context.CancelFunc
	done   chan struct{}
}

const rebuildTimeout = 10 * time.Minute

func startRebuild(repo string, project Project) {
	buildMu.Lock()
	prev := builds[repo]
	if prev != nil {
		log.Printf("[%s] superseding in-flight rebuild", repo)
		prev.cancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), rebuildTimeout)
	b := &build{cancel: cancel, done: make(chan struct{})}
	builds[repo] = b
	buildMu.Unlock()

	go func() {
		defer close(b.done)
		defer cancel()

		if prev != nil {
			<-prev.done
		}

		if ctx.Err() == nil {
			rebuild(ctx, repo, project)
		}

		buildMu.Lock()
		if builds[repo] == b {
			delete(builds, repo)
		}

		buildMu.Unlock()
	}()
}

func rebuild(ctx context.Context, repo string, project Project) {
	dir := expandHome(project.Path)
	log.Printf("[%s] rebuild starting in %s", repo, dir)

	target := "@{u}"
	if project.Branch != "" {
		target = "origin/" + project.Branch
	}

	steps := [][]string{
		{"git", "fetch", "origin"},
		{"git", "reset", "--hard", target},
		{"docker", "compose", "up", "-d", "--build"},
	}

	for _, step := range steps {
		log.Printf("[%s] $ %s", repo, strings.Join(step, " "))

		cmd := exec.CommandContext(ctx, step[0], step[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if s := strings.TrimSpace(string(out)); s != "" {
			log.Printf("[%s] %s", repo, s)
		}

		if err != nil {
			switch ctx.Err() {
			case context.DeadlineExceeded:
				log.Printf("[%s] rebuild timed out after %s — aborting", repo, rebuildTimeout)
			case context.Canceled:
				log.Printf("[%s] rebuild cancelled — superseded by a newer push", repo)
			default:
				log.Printf("[%s] command failed: %v — aborting rebuild", repo, err)
			}

			return
		}
	}

	log.Printf("[%s] rebuild complete", repo)
}

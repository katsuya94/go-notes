package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type result struct {
	title string
	count int
}

type Results struct {
	results []result
	m       map[string]*result
}

func NewResults() Results {
	return Results{nil, map[string]*result{}}
}

func (r *Results) Add(title string) {
	var (
		res *result
		ok  bool
	)
	res, ok = r.m[title]
	if !ok {
		r.results = append(r.results, result{title, 0})
		res = &r.results[len(r.results)-1]
		r.m[title] = res
	}
	res.count++
}

func (r *Results) Len() int {
	return len(r.results)
}

func (r *Results) Less(i, j int) bool {
	return r.results[i].count < r.results[j].count
}

func (r *Results) Swap(i, j int) {
	r.results[i], r.results[j] = r.results[j], r.results[i]
}

func (r *Results) Sorted() []string {
	sort.Sort(r)
	titles := make([]string, len(r.results))
	for i := range titles {
		titles[i] = r.results[i].title
	}
	return titles
}

type SearchManagerOptions struct {
	Selection chan<- string
}

type SearchManager struct {
	Options        SearchManagerOptions
	notesDirectory string
	query          []rune
	results        []string
	queryTrigger   *Trigger
	trigger        *Trigger
	mutex          *sync.RWMutex
}

func NewSearchManager(options SearchManagerOptions) *SearchManager {
	return &SearchManager{options, "", nil, nil, NewTrigger(), NewTrigger(), &sync.RWMutex{}}
}

func (sm *SearchManager) Client() *SearchClient {
	return &SearchClient{sm}
}

func (sm *SearchManager) notify() {
	Logger.Print("Notify Search")
	sm.trigger.Notify()
}

func (sm *SearchManager) notifyQuery() {
	sm.queryTrigger.Notify()
}

func (sm *SearchManager) doSearch() error {
	Logger.Print("Searching")
	var (
		query   = string(sm.query)
		titles  = []string{}
		results = NewResults()
		err     error
	)
	walkFunc := func(p string, info os.FileInfo, err error) error {
		if p == sm.notesDirectory {
			if err != nil {
				return err
			}
			return nil
		}
		if err != nil {
			Logger.Print("Error while walking directory: ", err)
			return nil
		}
		if info.IsDir() {
			Logger.Print("Encountered nested directory")
			return filepath.SkipDir
		}
		base := path.Base(p)
		title := strings.TrimSuffix(base, ".txt")
		if title == base {
			Logger.Printf("Encountered malformed filename: %s", base)
			return nil
		}
		titles = append(titles, title)
		return nil
	}
	err = filepath.Walk(sm.notesDirectory, walkFunc)
	if err != nil {
		return err
	}

	Logger.Print("Found ", len(titles), " notes")

	cmd := exec.Command("grep", "-i", query)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return err
	}

	for _, title := range titles {
		_, err = fmt.Fprintf(stdin, "%s\n", title)
		if err != nil {
			return err
		}
	}
	err = stdin.Close()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		// TODO: handle multiple occurrences in
		results.Add(strings.TrimRight(scanner.Text(), "\n"))
	}

	err = cmd.Wait()
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			return err
		}
	} else if err != nil {
		return err
	}

	Logger.Print("Found ", results.Len(), " results")

	sm.mutex.Lock()
	sm.results = results.Sorted()
	sm.mutex.Unlock()
	sm.notify()
	return nil
}

func (sm *SearchManager) Start() error {
	if sm.Options.Selection == nil {
		return fmt.Errorf("no Selection")
	}
	var err error
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sm.notesDirectory = path.Join(home, "Notes")
	subscription := sm.queryTrigger.Subscribe()
	Logger.Print("Starting SearchManager")
	for {
		subscription.Wait()
		err = sm.doSearch()
		if err != nil {
			return err
		}
	}
}

type SearchClient struct {
	sm *SearchManager
}

func (sc *SearchClient) Query() string {
	return string(sc.sm.query)
}

func (sc *SearchClient) Results() []string {
	sc.sm.mutex.RLock()
	results := make([]string, len(sc.sm.results))
	copy(results, sc.sm.results)
	sc.sm.mutex.RUnlock()
	return results
}

func (sc *SearchClient) Append(c rune) {
	sc.sm.query = append(sc.sm.query, c)
	sc.sm.notify()
	sc.sm.notifyQuery()
}

func (sc *SearchClient) Backspace() {
	if len(sc.sm.query) == 0 {
		return
	}
	sc.sm.query = sc.sm.query[:len(sc.sm.query)-1]
	sc.sm.notify()
	sc.sm.notifyQuery()
}

func (sc *SearchClient) Select() {
	query := string(sc.sm.query)
	if query == "" {
		sc.sm.Options.Selection <- ""
		return
	}
	notePath := path.Join(sc.sm.notesDirectory, query+".txt")
	sc.sm.Options.Selection <- notePath
}

func (sc *SearchClient) Subscribe() Subscription {
	return sc.sm.trigger.Subscribe()
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/CaliDog/certstream-go"
	"github.com/fsnotify/fsnotify"
)

func loadScopeFile(scopeFile string) ([]string, error) {
	pf, err := os.Open(scopeFile)
	if err != nil {
		return nil, err
	}
	defer pf.Close()

	var scopes []string

	sc := bufio.NewScanner(pf)
	for sc.Scan() {
		scope := sc.Text()
		if scope != "" {
			// prefix domain with dot for easier pattern matching with HasSuffix()
			scopes = append(scopes, fmt.Sprintf(".%s", scope))
		}
	}

	return scopes, nil
}

func watchScopeFile(scopeFile string, scopeChan chan []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file watcher: %s\n", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(scopeFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding file to watcher: %s\n", err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// File was modified, reload the scope file
				scopes, err := loadScopeFile(scopeFile)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reloading scope file: %s\n", err)
				} else {
					scopeChan <- scopes
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "Error watching file: %s\n", err)
		}
	}
}

func main() {
	var scopeFile string
	var wildcardsOnly bool

	flag.StringVar(&scopeFile, "s", "", "Scope file")
	flag.BoolVar(&wildcardsOnly, "w", false, "Output wildcard domains only")

	flag.Parse()

	if scopeFile != "" {
		scopes, err := loadScopeFile(scopeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading scope file: %s\n", err)
			return
		}

		scopeChan := make(chan []string)

		go watchScopeFile(scopeFile, scopeChan)

		w := bufio.NewWriter(os.Stdout)
		defer w.Flush()

		stream, errStream := certstream.CertStreamEventStream(false)
		for {
			select {
			case jq := <-stream:
				domains, err := jq.ArrayOfStrings("data", "leaf_cert", "all_domains")
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error decoding json", err)
				} else {
					for _, domain := range domains {
						for _, scope := range scopes {
							if strings.HasSuffix(domain, scope) {
								if wildcardsOnly {
									if domain[0:2] == "*." {
										fmt.Fprintln(w, domain)
									}
								} else {
									fmt.Fprintln(w, strings.Replace(domain, "*.", "", 1))
								}
							}
						}
					}
				}
			case newScopes := <-scopeChan:
				fmt.Fprintf(os.Stderr, "Scope file change detected. Reloading.\n")
				scopes = newScopes
			case err := <-errStream:
				fmt.Fprintln(os.Stderr, "Stream error", err)
			}
		}
	} else {
		flag.Usage()
		return
	}
}

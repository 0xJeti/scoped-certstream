package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/CaliDog/certstream-go"
)

func main() {
	var scopeFile string
	var scopes []string
	var wildcardsOnly bool

	flag.StringVar(&scopeFile, "s", "", "Scope file")
	flag.BoolVar(&wildcardsOnly, "w", false, "Output wildcard domains only")

	flag.Parse()

	if scopeFile != "" {
		pf, err := os.Open(scopeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}

		sc := bufio.NewScanner(pf)

		for sc.Scan() {
			scope := sc.Text()
			if scope != "" {
				// prefix domain with dot for easier pattern matching with HasSuffix()
				scopes = append(scopes, fmt.Sprintf(".%s", scope))
			}
		}
	} else {
		flag.Usage()
		return
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	// The false flag specifies that we want heartbeat messages.
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
		case err := <-errStream:
			fmt.Fprintln(os.Stderr, "Stream error", err)
		}
	}
}

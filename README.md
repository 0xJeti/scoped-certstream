# Introduction
scoped-certstream is a small program that constantly retrieves a stream of domains from CertStream and prints them to standard output.

Only domains that match the scope are printed. The scope is defined in a flat file containing domain suffixes.
Following scope file:
```
redbull.com
tesla.com
```

will match domains e.g.:
```
test.redbull.com
test.test.redbull.com
test.tesla.com
```

but won't match:
```
www.temperedbull.com
```

# Installation
Standard `go` way:
```bash
go install github.com/0xJeti/scoped-certstream@latest
```
# Usage
```
Usage of scoped-certstream:
  -s string
        Scope file
  -w    Output wildcard domains only
```
With `-w` flag *scoped-certsream* will output only wildcard cert domains (prefixed with *.)

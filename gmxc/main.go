package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

var (
	delay    = flag.Duration("d", 0, "delay between updates")
	duration = flag.Duration("D", 0, "duration to output continually")

	pid   = flag.Int("p", 0, "process to inspect")
	pname = flag.String("n", "", "name of process to inspect")

	socketregex = regexp.MustCompile(`\.gmx\.[0-9]+\.[0-9]+`)
)

type conn struct {
	net.Conn
	*json.Decoder
	*json.Encoder
}

func listGmxProcesses(f func(file string, args interface{})) {
	dir, err := os.Open(os.TempDir())
	if err != nil {
		log.Fatalf("unable to open %s: %v", os.TempDir(), err)
	}
	pids, err := dir.Readdirnames(0)
	if err != nil {
		log.Fatalf("unable to read pids: %v", err)
	}
	for _, pid := range pids {
		if socketregex.MatchString(pid) {
			c, err := dial(filepath.Join(os.TempDir(), pid))
			if err != nil {
				continue
			}
			defer c.Close()
			c.Encode([]string{"os.args"})
			var result = make(map[string]interface{})
			if err := c.Decode(&result); err != nil {
				log.Printf("unable to decode response from %s: %v", pid, err)
				continue
			}
			if args, ok := result["os.args"]; ok {
				f(pid, args)
			}
		}
	}
}

func findGmxProcess(pname string) int {
	var found int
	listGmxProcesses(func(file string, args interface{}) {
		if argslist, ok := args.([]interface{}); ok && len(argslist) >= 1 {
			name, ok := argslist[0].(string)
			if ok {
				if filepath.Base(name) == pname {
					str_pid := file[5 : len(file)-2] // ".gmx.####.#"
					numeric_pid, err := strconv.Atoi(str_pid)
					if err == nil {
						if found == 0 {
							fmt.Printf("Using %s\t%v\n", name, args)
							found = numeric_pid
						} else if found > 0 {
							fmt.Printf("Ambiguous situation. Both %d and %d could be %s. Use -p option\n", found, numeric_pid, pname)
							found = -1
						}
					}
				}
			}
		}
	})
	if found > 0 {
		return found
	}
	return 0
}

// fetchKeys returns all the registered keys from the process.
func fetchKeys(c *conn) []string {
	// retrieve list of registered keys
	if err := c.Encode([]string{"keys"}); err != nil {
		log.Fatalf("unable to send keys request to process: %v", err)
	}
	var result = make(map[string][]string)
	if err := c.Decode(&result); err != nil {
		log.Fatalf("unable to decode keys response: %v", err)
	}
	keys, ok := result["keys"]
	if !ok {
		log.Fatalf("gmx server did not return a keys list")
	}
	sort.Strings(keys)
	return keys
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintf(os.Stderr, "  %s [OPTIONS] [REGEX PATTERN]*\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *pid == 0 && *pname != "" {
		*pid = findGmxProcess(*pname)
	}
	if *pid == 0 {
		listGmxProcesses(func(name string, args interface{}) { fmt.Printf("%s\t%v\n", name, args) })
		return
	}
	c, err := dial(filepath.Join(os.TempDir(), fmt.Sprintf(".gmx.%d.0", *pid)))
	if err != nil {
		log.Fatalf("unable to connect to process %d: %v", *pid, err)
	}
	defer c.Close()

	// have JSON decoder preserve large integers as themselves, rather than convert to float64 and loose precision
	c.UseNumber()

	// match flag.Args() as regexps
	registeredKeys := fetchKeys(c)
	var keys []string
	if len(flag.Args()) == 0 {
		// no patterns? then you get everything
		keys = registeredKeys
	} else {
		for _, a := range flag.Args() {
			r, err := regexp.Compile(a)
			if err != nil {
				log.Fatalf("unable to compile regex %v: %v", a, err)
			}
			for _, k := range registeredKeys {
				if r.MatchString(k) {
					keys = append(keys, k)
				}
			}
		}
	}

	deadline := time.Now().Add(*duration)
	for {
		if err := c.Encode(keys); err != nil {
			log.Fatalf("unable to send request to process: %v", err)
		}
		var result = make(map[string]interface{})
		if err := c.Decode(&result); err != nil {
			log.Fatalf("unable to decode response: %v", err)
		}
		for _, k := range keys {
			if v, ok := result[k]; ok {
				fmt.Printf("%s: %v\n", k, v)
			}
		}
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(*delay)
		fmt.Println("----") // separator line
	}
}

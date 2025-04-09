package main

/*
retry
	a dumb application for dumb problems.

	Retries a command until that program exits with a success code.
	Absolutely no handling of any other state. Be careful.
	Retries have a backoff of (last_wait * log2(retry_i).
	Larger values of -t may wait for a long time.
*/

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var retries = flag.Int("t", 40, "how many times to retry.")
var fixed_interval = flag.Int("interval", 0, "use a fixed interval between retries (default mode is log2 backoff)")
var interval_unit = flag.String("unit", "seconds", "interval unit (required --interval)")
var verbose = flag.Bool("verbose", false, "verbose output")
var identifier = flag.String("id", "", "log identifier (default is the retry command and args)")
var attempts int = 1
var rwlock = sync.RWMutex{}

func updateSpew(arguments string) {
	for {
		rwlock.RLock()
		os.Stderr.WriteString(fmt.Sprintf("retry: %s has %d retries remaining...\n", arguments, *retries-attempts))
		rwlock.RUnlock()
		time.Sleep(time.Duration(300) * time.Second)
	}

}
func main() {

	flag.Parse()

	if len(flag.Args()) <= 0 {
		log.Println("found no arguments for a program to run. Exiting...")
		os.Exit(1)
	}
	var interval_multiplier int
	switch *interval_unit {
	case "h":
		{
			interval_multiplier = 60000
		}
	case "s":
		{
			interval_multiplier = 1000
		}
	case "ms":
		{
			interval_multiplier = 100
		}
	default:
		{
			// assume seconds
			interval_multiplier = 1000
		}
	}
	var errorsSeen = make(map[string]int)
	var interval int
	if *fixed_interval > 0 {
		interval = *fixed_interval * interval_multiplier
	} else {
		interval = 100 // start at 100ms for backoff version
	}
	var logID string
	if *identifier == "" {
		logID = strings.Join(flag.Args(), " ")
	} else {
		logID = "id '" + *identifier + "'"
	}
	go updateSpew(logID)
	for i := 2; i < *retries+2; i++ {
		var cmd = exec.Command(flag.Arg(0), flag.Args()[1:]...)
		output, err := cmd.CombinedOutput()
		if err == nil {
			os.Stdout.WriteString(fmt.Sprintf("%s\n", output))
			os.Exit(0)
		}
		// otherwise, there was a problem.
		if *fixed_interval > 0 {
			// interval doesn't change :)
		} else {
			interval *= int(math.Log2(float64(i)))
		}
		strOutput := fmt.Sprint(output)
		hasOutput := strings.TrimSpace(strOutput) == ""
		_, seen := errorsSeen[strOutput]
		logOutput := ""
		if *verbose {
			logOutput += fmt.Sprintf("retry: Error %s, retrying after %.2e seconds.\n", err.Error(), time.Duration(interval).Seconds())
		}
		if !seen && hasOutput {
			errorsSeen[strOutput] = 1
			logOutput += fmt.Sprintf("retry: %s has new output:\n%s\n", logID, output)
		}
		if logOutput != "" {
			os.Stderr.WriteString(logOutput)
		}
		rwlock.Lock()
		attempts += 1
		rwlock.Unlock()

		time.Sleep(time.Millisecond * time.Duration(interval))
	}
	log.Printf("retry: %d attempts failed for %s\n", *retries, strings.Join(flag.Args(), ","))
	os.Exit(1)
}

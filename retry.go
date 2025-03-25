package main

/*
retry
	a dumb application for dumb problems.

	Retries a command until that program exits with a success code.
	Absolutely no handling of state. Be careful.
	Retries have a backoff of (last_wait * log2(retry_i).
	Larger values of -t may wait for a long time.
*/

import (
	"flag"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"
)

var retries = flag.Int("t", 40, "how many times to retry.")

func main() {
	flag.Parse()

	if len(flag.Args()) <= 0 {
		log.Println("found no arguments for a program to run. Exiting...")
		os.Exit(1)
	}
	backoff := 100
	for i := 2; i < *retries+2; i++ {
		var cmd = exec.Command(flag.Arg(0), flag.Args()[1:]...)
		output, err := cmd.CombinedOutput()
		if err == nil {
			log.Printf("%s\n", output)
			os.Exit(0)
		}
		// otherwise, there was a problem.
		backoff *= int(math.Log2(float64(i)))
		log.Printf("Error %s, retrying after %.2e milliseconds. Output:\n", err.Error(), float64(backoff)*1000)
		log.Printf("%s\n", output)
		time.Sleep(time.Millisecond * time.Duration(backoff))
	}
	log.Printf("Process retry attempts failed for %s\n", strings.Join(flag.Args(), ","))
	os.Exit(1)
	//os.StartProcess()
}

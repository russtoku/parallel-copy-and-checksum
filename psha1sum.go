package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
)

var maxWorkers int = 30

// psha1sum computes the SHA1 sum of files in the top-level of a
// directory using up to 30 concurrent goroutines. The default
// number of goroutines is 10.
func main() {
	var numWorkers = flag.Int("w", 10, "number of workers")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalf("usage: %s [ -w ] dir", os.Args[0])
	}

	log.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0))
	log.Println("NumCPU:", runtime.NumCPU())

	dir := flag.Args()[0]

	if *numWorkers > 30 {
		log.Println(*numWorkers, "specifed; max is", maxWorkers)
		*numWorkers = maxWorkers
	}

	log.Println("number of workers:", *numWorkers)

	todo := make(chan string)
	results := make(chan string, *numWorkers)
	var wg sync.WaitGroup

	files := filesIn(dir)

	// queue up files to work on
	go func(fs []string) {
		for _, f := range fs {
			todo <- f
		}
		close(todo)
	}(files)

	// start the workers
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for f := range todo {
				fpath := fmt.Sprintf("%s/%s", dir, f)
				results <- fmt.Sprintf("%s: %s", f, sha1Sum(fpath))
			}
		}(i + 1)
	}

	// when workers are done, close the results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// get the results
	for line := range results {
		fmt.Println(line)
	}
}

// Return a list of files in the top level of dir
func filesIn(dir string) []string {
	var fs []string

	d, _ := os.Open(dir)
	names, err := d.Readdirnames(0)
	if err != nil {
		log.Fatal("filesIn: ", err)
	}

	for _, name := range names {
		finfo, err := os.Lstat(dir + "/" + name)
		if err != nil {
			log.Fatal("filesIn: ", err)
		}
		if finfo.Mode().IsRegular() {
			fs = append(fs, finfo.Name())
		}
	}
	return fs
}

// Return the SHA1 sum of filename as a string
func sha1Sum(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal("sha1Sum: ", err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Println("sha1Sum:", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

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

// pcopy copies files from the top-level of a source directory to a
// destination directory using up to 30 concurrent goroutines. The
// default number of goroutines is 10. The destination  directory must
// exist.
func main() {
	var numWorkers = flag.Int("w", 10, "number of workers")
	flag.Parse()

	if len(flag.Args()) < 2 {
		log.Fatalf("usage: %s [ -w ] srcDir destDir", os.Args[0])
	}

	log.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0))
	log.Println("NumCPU:", runtime.NumCPU())

	srcDir := flag.Args()[0]
	destDir := flag.Args()[1]

	if *numWorkers > 30 {
		log.Println(*numWorkers, "specifed; max is", maxWorkers)
		*numWorkers = maxWorkers
	}

	log.Println("number of workers:", *numWorkers)

	todo := make(chan string)
	results := make(chan string, *numWorkers)
	var wg sync.WaitGroup

	if !isThere(destDir) {
		log.Fatal("Can't find directory ", destDir)
	}

	files := filesIn(srcDir)

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
				srcName := fmt.Sprintf("%s/%s", srcDir, f)
				destName := fmt.Sprintf("%s/%s", destDir, f)
				numBytes, chksum := copyAndSha1Sum(srcName, destName)
				results <- fmt.Sprintf("%s %d %s", f, numBytes, chksum)
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

// Check for existence of dir
func isThere(dir string) bool {
	if _, err := os.Open(dir); err != nil {
		return false
	}
	return true
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

// Copy file from src to dest and compute SHA1 sum of bytes copied
// Returns the number of bytes written and the checksum
func copyAndSha1Sum(srcName string, destName string) (int64, string) {
	src, err := os.Open(srcName)
	if err != nil {
		log.Fatal("copyAndSha1Sum: ", err)
	}
	defer src.Close()

	dest, err := os.Create(destName)
	if err != nil {
		log.Fatal("copyAndSha1Sum: ", err)
	}
	defer dest.Close()

	h := sha1.New()
	tee := io.MultiWriter(dest, h)
	bytesWritten, err := io.Copy(tee, src)
	if err != nil {
		log.Fatal("copyAndSha1Sum: ", err)
	}

	return bytesWritten, fmt.Sprintf("%x", h.Sum(nil))
}

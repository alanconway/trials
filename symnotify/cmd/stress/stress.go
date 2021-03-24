package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	dir       = flag.String("dir", "/tmp/symnotify", "directory for links")
	runTime   = flag.Duration("time", time.Duration(10*time.Second), "time to run test")
	fileCount = flag.Int64("files", 1000, "number of files")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %v [FLAGS]
Create FILES symlinks under DIR to files under DIR/files.
Randomly write and rotate the files for TIME.
`)
	}
	var (
		fileCount = 100
		links     = *dir
		files     = join(*dir, "files")
	)
	defer os.RemoveAll(*dir)
	fatalErr(os.MkdirAll(links, os.ModePerm))
	fatalErr(os.MkdirAll(files, os.ModePerm))
	deadline := time.Now().Add(*runTime)
	g := sync.WaitGroup{}
	for i := 0; i < fileCount; i++ {
		g.Add(1)
		name := fmt.Sprintf("test%09v.log", i)
		f, err := os.Create(join(files, name))
		fatalErr(err)
		fatalErr(f.Close())
		fatalErr(os.Symlink(join(files, name), join(links, name)))
		go func() {
			defer g.Done()
			var writes, rotates int
			for time.Now().Before(deadline) {
				x := rand.Intn(100)
				switch {
				case x < 5: // rotate 5% of the time.
					rotates++
					f, err := os.Create(join(files, name+"_rotate"))
					fatalErr(err)
					fatalErr(f.Close())
					fatalErr(os.Rename(join(files, name+"_rotate"), join(files, name)))
				default: // Write data 90% of the time.
					f, err := os.OpenFile(join(files, name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
					fatalErr(err)
					_, err = f.Write([]byte{byte('X')})
					fatalErr(err)
					fatalErr(f.Close())
					writes++
				}
				time.Sleep(time.Duration(x / 10 * int(time.Millisecond)))
			}
			log.Printf("%v: writes: %v rotates: %v", name, writes, rotates)
		}()
	}
	g.Wait()
}

func fatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var join = filepath.Join

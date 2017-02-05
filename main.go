package main

/* requirements:
1) avoid concurrent downloads of the same file
2) make progress towards empty
3) cURL compatible (-XFOO does not become GET upon 303 See Other)
4) resumable transfers via curl & HTTP 1.1 range
5) walk subdirs?
*/

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// A popFileSystem is a http.FileSystem with an eventually-destructive
// non-standard POP method.  Files in transit via the POP method are
// stashed into a nearby subdirectory as to not appear to concurrent
// clients.
type popFileSystem struct {
	root http.Dir
}

// A popSubdir is where files are moved to conceal them from
// subsequent or nearly-concurrent Readdir()s.
const popSubdir = ".pop"

func (fs popFileSystem) Stash(name string) error {
	src := filepath.Join(string(fs.root), name)
	dst := filepath.Join(filepath.Dir(src), popSubdir, filepath.Base(src))
	err := os.Mkdir(filepath.Dir(dst), 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return os.Rename(src, dst)
}

// Pop() selects a suitable file and stashes it to avoid simultaneous
// downloads.
func (fs popFileSystem) Pop(prefix string) (string, error) {
	f, err := fs.root.Open(prefix)
	if err != nil {
		return "", err
	}
	fileinfos, err := f.Readdir(-1)
	if err != nil {
		return "", err
	}
	for _, fileinfo := range fileinfos {
		// skip directories
		if fileinfo.IsDir() {
			continue
		}

		name := fileinfo.Name()
		fullName := filepath.Join(prefix, name)

		// skip dotfiles (such as popSubdir)
		if strings.HasPrefix(name, ".") {
			continue
		}

		// hide file in nearby subdirectory
		err = fs.Stash(fullName)
		if os.IsExist(err) {
			continue
		} else if err != nil {
			return "", err
		}

		// reference file via its unstashed path, and finish
		return fullName, nil
	}

	// all files are dotfiles or unstashable
	return "", os.ErrNotExist
}

// Open() accesses a (possibly stashed) file or directory.
func (fs popFileSystem) Open(name string) (http.File, error) {
	f, err := fs.root.Open(name)
	if err != nil {
		// read stashed files (if present)
		stashedName := filepath.Join(filepath.Dir(name), popSubdir, filepath.Base(name))
		return fs.root.Open(stashedName)
	}
	return f, nil
}

var port int

func init() {
	flag.IntVar(&port, "port", 8666, "port")
	flag.Parse()
}

func main() {
	dir := http.Dir(".")
	fs := popFileSystem{dir}
	fsrv := http.FileServer(fs)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POP":
			name, err := fs.Pop(r.URL.Path)
			if err != nil {
				http.Error(w, "500 internal server error", http.StatusInternalServerError)
				log.Println(err)
				return
			}

			r, err = http.NewRequest("GET", name, nil)
			if err != nil {
				http.Error(w, "500 internal server error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition",
				fmt.Sprintf("attachment; filename=%#v", name))
		}
		log.Println(r.Method, r.URL.Path)
		fsrv.ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

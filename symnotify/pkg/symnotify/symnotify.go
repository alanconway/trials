// package symnotify provides a file system watcher that notifies events for symlink targets.
//
package symnotify

import (
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Event = fsnotify.Event
type Op = fsnotify.Op

const (
	Create Op = fsnotify.Create
	Write     = fsnotify.Write
	Remove    = fsnotify.Remove
	Rename    = fsnotify.Rename
	Chmod     = fsnotify.Chmod
)

// Watcher is like fsnotify.Watcher but also notifies on changes to symlink targets
type Watcher struct {
	watcher *fsnotify.Watcher
	added   map[string]bool
}

func NewWatcher() (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	return &Watcher{watcher: w, added: map[string]bool{}}, err
}

// Event returns the next event.
func (w *Watcher) Event() (e Event, err error) {
	return w.EventTimeout(time.Duration(math.MaxInt64))
}

// EventTimeout returns the next event or os.ErrDeadlineExceeded if timeout is exceeded.
func (w *Watcher) EventTimeout(timeout time.Duration) (e Event, err error) {
	var ok bool
	select {
	case e, ok = <-w.watcher.Events:
	case err, ok = <-w.watcher.Errors:
	case <-time.After(timeout):
		return Event{}, os.ErrDeadlineExceeded
	}
	switch {
	case !ok:
		return Event{}, io.EOF
	case e.Op == Create:
		if info, err := os.Lstat(e.Name); err == nil {
			w.addIfSymlink(info, e.Name)
		}
	case e.Op == Remove && !w.added[e.Name]:
		w.watcher.Remove(e.Name)
	}
	return e, err
}

// Add a file to the watcher
func (w *Watcher) Add(name string) error {
	if err := w.watcher.Add(name); err != nil {
		return err
	}
	w.added[name] = true // Explicitly added, don't auto-Remove

	// Scan directories for existing symlinks, we wont' get a Create for those.
	if infos, err := ioutil.ReadDir(name); err == nil {
		for _, stat := range infos {
			w.addIfSymlink(stat, filepath.Join(name, stat.Name()))
		}
	}
	return nil
}

// addIfSymlink adds name if info is a symlink.
//
// Ordinary files in a watched directory are watched automatically.
// Symlinks are watched for create/destroy but the symlink itself never changes.
// To get events for the target, we must directly watch the symlink name.
func (w *Watcher) addIfSymlink(info os.FileInfo, name string) {
	if info.Mode()&os.ModeSymlink == os.ModeSymlink {
		_ = w.watcher.Add(name)
	}
}

// Remove name from watcher
func (w *Watcher) Remove(name string) error {
	delete(w.added, name)
	return w.watcher.Remove(name)
}

// Close watcher
func (w *Watcher) Close() error { return w.watcher.Close() }

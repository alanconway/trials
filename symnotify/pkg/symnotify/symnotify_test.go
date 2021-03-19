package symnotify

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	// Utility functions to simplify the test
	join := func(dir, name string) string { return filepath.Join(dir, name) }
	link := func(oldname, newname string) {
		t.Helper()
		require.NoError(os.Symlink(oldname, newname))
	}
	create := func(name string) *os.File {
		t.Helper()
		f, err := os.Create(name)
		require.NoError(err)
		t.Cleanup(func() { f.Close() })
		return f
	}

	dir := t.TempDir()  // Watched directory
	dir2 := t.TempDir() // Directory for symlink targets

	// Add file and link before starting to watch
	file1 := create(join(dir, "file1"))
	target1 := create(join(dir2, "target1"))
	link(join(dir2, "target1"), join(dir, "link1"))

	// Start watching dir
	w, err := NewWatcher()
	require.NoError(err)
	defer w.Close()
	require.NoError(w.Add(dir))

	// Watcher should not have any events yet
	_, err = w.EventTimeout(time.Duration(0))
	require.Equal(err, os.ErrDeadlineExceeded)

	// Function to get next event
	event := func() Event {
		t.Helper()
		e, err := w.EventTimeout(time.Second)
		require.NoError(err)
		return e
	}

	// Create file and link after starting watch
	file2 := create(join(dir, "file2"))
	assert.Equal(event(), Event{Name: join(dir, "file2"), Op: Create})
	target2 := create(join(dir2, "target2"))
	link(join(dir2, "target2"), join(dir, "link2"))
	assert.Equal(event(), Event{Name: join(dir, "link2"), Op: Create})

	// Write to real files, check events.
	file1.Write([]byte("hello"))
	assert.Equal(event(), Event{Name: join(dir, "file1"), Op: Write})
	file1.Truncate(0)
	assert.Equal(event(), Event{Name: join(dir, "file1"), Op: Write})
	file2.Write([]byte("hello"))
	assert.Equal(event(), Event{Name: join(dir, "file2"), Op: Write})

	// Write to targets, check events on links.
	target1.Write([]byte("hello"))
	assert.Equal(event(), Event{Name: join(dir, "link1"), Op: Write})
	target2.Write([]byte("hello"))
	assert.Equal(event(), Event{Name: join(dir, "link2"), Op: Write})
	target2.Chmod(0444)
	assert.Equal(event(), Event{Name: join(dir, "link2"), Op: Chmod})

	// Rename and delete
	assert.NoError(os.Rename(join(dir, "file1"), join(dir, "newfile1")))
	assert.Equal(event(), Event{Name: join(dir, "file1"), Op: Rename})
	assert.Equal(event(), Event{Name: join(dir, "newfile1"), Op: Create})
	assert.NoError(os.Remove(join(dir, "link1")))
	assert.Equal(event(), Event{Name: join(dir, "link1"), Op: Remove})

	// Removing a link target is a "Chmod" event, not a remove since the link is still there.
	assert.NoError(os.Remove(join(dir2, "target2")))
	assert.Equal(event(), Event{Name: join(dir, "link2"), Op: Chmod})

	// Watcher should not have any events
	_, err = w.EventTimeout(time.Duration(0))
	require.Equal(err, os.ErrDeadlineExceeded)

}

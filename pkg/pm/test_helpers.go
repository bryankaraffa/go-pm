package pm

import (
	"os"
	"strings"
)

// MockFileSystem is a mock implementation of FileSystem for testing
type MockFileSystem struct {
	files map[string][]byte
	dirs  map[string]bool
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

func (fs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if content, exists := fs.files[path]; exists {
		return content, nil
	}
	return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}

func (fs *MockFileSystem) WriteFile(path string, content []byte) error {
	fs.files[path] = content
	return nil
}

func (fs *MockFileSystem) FileExists(path string) bool {
	_, exists := fs.files[path]
	return exists
}

func (fs *MockFileSystem) DirectoryExists(path string) bool {
	return fs.dirs[path]
}

func (fs *MockFileSystem) CreateDirectory(path string) error {
	fs.dirs[path] = true
	return nil
}

func (fs *MockFileSystem) ListDirectories(path string) ([]string, error) {
	var dirs []string
	for dir := range fs.dirs {
		if strings.HasPrefix(dir, path+"/") || dir == path {
			// Extract just the directory name, not the full path
			if dir != path {
				relPath := strings.TrimPrefix(dir, path+"/")
				if !strings.Contains(relPath, "/") {
					dirs = append(dirs, relPath)
				}
			}
		}
	}
	return dirs, nil
}

func (fs *MockFileSystem) CopyFile(src, dst string) error {
	if content, exists := fs.files[src]; exists {
		fs.files[dst] = content
		return nil
	}
	return &os.PathError{Op: "open", Path: src, Err: os.ErrNotExist}
}

func (fs *MockFileSystem) ListFiles(path string) ([]string, error) {
	var files []string
	for file := range fs.files {
		if strings.HasPrefix(file, path) {
			files = append(files, file)
		}
	}
	return files, nil
}

func (fs *MockFileSystem) MoveDirectory(src, dst string) error {
	// Mark destination as existing and remove source
	fs.dirs[dst] = true
	delete(fs.dirs, src)
	return nil
}

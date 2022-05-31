package diff

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pkg/diff"
	"github.com/pkg/diff/write"
)

const (
	homeSymbol = '~'
)

var colorWarning = color.New(color.FgYellow)

// Diff src with dst (src and dst can have home symbol)
func Diff(src, dst string) error {
	// Trim home symbol first to make sure no ~ in path
	src, err := trimHomeSymbol(src)
	if err != nil {
		return fmt.Errorf("failed to trim home symbol src [%s]", src)
	}

	dst, err = trimHomeSymbol(dst)
	if err != nil {
		return fmt.Errorf("failed to trim home symbol dst [%s]", dst)
	}

	return diffRaw(src, dst)
}

// Diff src with dst
func diffRaw(src, dst string) error {
	srcFileInfo, err := os.Stat(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			colorWarning.Printf("src [%s] not exist\n", src)
			return nil
		}

		return fmt.Errorf("failed to stat src [%s]: %w", src, err)
	}

	dstFileInfo, err := os.Stat(dst)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			colorWarning.Printf("dst [%s] not exist\n", dst)
			return nil
		}

		return fmt.Errorf("failed to stat dst [%s]: %w", dst, err)
	}

	// Both is dir
	if srcFileInfo.IsDir() && dstFileInfo.IsDir() {
		if err := diffDir(src, dst); err != nil {
			return fmt.Errorf("failed to diff dir src [%s] dst [%s]: %w", src, dst, err)
		}

		return nil
	}

	// Both is file
	if !srcFileInfo.IsDir() && !dstFileInfo.IsDir() {
		if err := diffFile(src, dst); err != nil {
			return fmt.Errorf("failed to diff file src [%s] dst [%s]: %w", src, dst, err)
		}

		return nil
	}

	colorWarning.Printf("src [%s] dst [%s] is not same type file or same type dir\n", src, dst)
	return nil
}

func diffDir(src, dst string) error {
	// Read dir into arr
	srcDirEntries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read dir [%s]: %w", src, err)
	}

	dstDirEntries, err := os.ReadDir(dst)
	if err != nil {
		return fmt.Errorf("failed to read dir [%s]: %w", dst, err)
	}

	// Convert arr to map
	mSrcEntries := make(map[string]fs.DirEntry)
	for _, entry := range srcDirEntries {
		mSrcEntries[entry.Name()] = entry
	}

	mDstEntries := make(map[string]fs.DirEntry)
	for _, entry := range dstDirEntries {
		mDstEntries[entry.Name()] = entry
	}

	// Find entry exist in both src and dst
	mBothExistEntries := make(map[string]fs.DirEntry)
	for name, entry := range mSrcEntries {
		if _, ok := mDstEntries[name]; ok {
			mBothExistEntries[name] = entry
		} else {
			colorWarning.Printf("src [%s] not exist in dst\n", filepath.Join(src, name))
		}
	}

	for entry := range mBothExistEntries {
		joinedSrc := filepath.Join(src, entry)
		joinedDst := filepath.Join(dst, entry)
		if err := diffRaw(joinedSrc, joinedDst); err != nil {
			return fmt.Errorf("failed to diff raw src [%s] dst [%s]: %w", src, dst, err)
		}
	}

	return nil
}

func diffFile(src, dst string) error {
	if err := diff.Text(src, dst, nil, nil, os.Stdout, write.TerminalColor()); err != nil {
		return fmt.Errorf("failed to diff text src [%s] dst [%s]: %w", src, dst, err)
	}

	return nil
}

// trimHomeSymbol replace ~ with full path
// Copy from https://github.com/make-go-great/copy-go
// https://stackoverflow.com/a/17609894
func trimHomeSymbol(path string) (string, error) {
	if path == "" || path[0] != homeSymbol {
		return path, nil
	}

	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	newPath := filepath.Join(currentUser.HomeDir, path[1:])
	return newPath, nil
}

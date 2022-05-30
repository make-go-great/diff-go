package diff

import (
	"errors"
	"fmt"
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

// Diff src with dst
func Diff(src, dst string) error {
	// Trim home symbol first to make sure no ~ in path
	src, err := trimHomeSymbol(src)
	if err != nil {
		return fmt.Errorf("failed to trim ~ for src [%s]", src)
	}

	dst, err = trimHomeSymbol(dst)
	if err != nil {
		return fmt.Errorf("failed to trim ~ for dst [%s]", dst)
	}

	srcFileInfo, err := os.Stat(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			colorWarning.Printf("src [%s] not exist\n", src)
			return nil
		}

		return fmt.Errorf("failed to stat src [%s]: %w", src, err)
	}

	dstFileInfo, err := os.Stat(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			colorWarning.Printf("dst [%s] not exist\n", src)
			return nil
		}

		return fmt.Errorf("failed to stat dst [%s]: %w", dst, err)
	}

	// Both is directory
	if srcFileInfo.IsDir() && dstFileInfo.IsDir() {
		if err := diffDir(src, dst); err != nil {
			return fmt.Errorf("failed to diff dir src [%s] dst [%s]: %w", src, dst, err)
		}
	}

	// Both is file
	if !srcFileInfo.IsDir() && !dstFileInfo.IsDir() {
		if err := diffFile(src, dst); err != nil {
			return fmt.Errorf("failed to diff file src [%s] dst [%s]: %w", src, dst, err)
		}
	}

	colorWarning.Printf("src %s dst %s is not same type file or same type directory\n", src, dst)
	return nil
}

func diffDir(src, dst string) error {
	srcDirEntries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read dir [%s]: %w", src, err)
	}

	dstDirEntries, err := os.ReadDir(dst)
	if err != nil {
		return fmt.Errorf("failed to read dir [%s]: %w", dst, err)
	}

	mSrcEntries := make(map[string]struct{})
	for _, entry := range srcDirEntries {
		mSrcEntries[filepath.Join(src, entry.Name())] = struct{}{}
	}

	mDstEntries := make(map[string]struct{})
	for _, entry := range dstDirEntries {
		mDstEntries[filepath.Join(dst, entry.Name())] = struct{}{}
	}

	mBothExistEntries := make(map[string]struct{})

	for entry := range mSrcEntries {
		if _, ok := mDstEntries[entry]; ok {
			mBothExistEntries[entry] = struct{}{}
		} else {
			colorWarning.Printf("src [%s] not exist in dst\n", entry)
		}
	}

	for entry := range mBothExistEntries {
		if err := Diff(entry, entry); err != nil {
			return fmt.Errorf("failed to diff src [%s] dst [%s]: %w", src, dst, err)
		}
	}

	return nil
}

func diffFile(src, dst string) error {
	if err := diff.Text(src, dst, nil, nil, os.Stdout, write.TerminalColor()); err != nil {
		return fmt.Errorf("failed to diff text src %s dst %s:%w", src, dst, err)
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

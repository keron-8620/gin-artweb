package archive

import (
	"archive/tar"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessTarEntryAllowsSiblingSymlink(t *testing.T) {
	root := t.TempDir()
	linkDir := filepath.Join(root, "legal", "java.desktop")
	if err := os.MkdirAll(linkDir, 0755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(linkDir, "COPYRIGHT")
	header := &tar.Header{
		Name:     "legal/java.desktop/COPYRIGHT",
		Linkname: "../java.base/COPYRIGHT",
		Typeflag: tar.TypeSymlink,
	}

	if err := processTarEntry(target, root, header, tar.NewReader(strings.NewReader("")), DefaultArchiveOptions); err != nil {
		t.Fatalf("expected sibling symlink to be allowed: %v", err)
	}
	link, err := os.Readlink(target)
	if err != nil {
		t.Fatal(err)
	}
	if link != header.Linkname {
		t.Fatalf("symlink target = %q, want %q", link, header.Linkname)
	}
}

func TestProcessTarEntryRejectsSymlinkOutsideRoot(t *testing.T) {
	root := t.TempDir()
	linkDir := filepath.Join(root, "legal", "java.desktop")
	if err := os.MkdirAll(linkDir, 0755); err != nil {
		t.Fatal(err)
	}
	header := &tar.Header{
		Name:     "legal/java.desktop/COPYRIGHT",
		Linkname: "../../../outside/COPYRIGHT",
		Typeflag: tar.TypeSymlink,
	}

	err := processTarEntry(filepath.Join(linkDir, "COPYRIGHT"), root, header, tar.NewReader(strings.NewReader("")), DefaultArchiveOptions)
	if err == nil {
		t.Fatal("expected symlink outside extraction root to be rejected")
	}
}

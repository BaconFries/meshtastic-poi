package source_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/BaconFries/meshtastic-poi/internal/downloader"
	"github.com/BaconFries/meshtastic-poi/internal/providers/source"
)

func TestReadLocalFile(t *testing.T) {
	path := filepath.Join("..", "..", "..", "testdata", "sample.gpx")
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	client := downloader.New("")
	data, err := source.Read(context.Background(), client, abs)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected file content")
	}

	fileURL := "file://" + abs
	data2, err := source.Read(context.Background(), client, fileURL)
	if err != nil {
		t.Fatal(err)
	}
	if len(data2) != len(data) {
		t.Fatal("file:// read mismatch")
	}
}

func TestReadFileURL(t *testing.T) {
	f, err := os.CreateTemp("", "meshtastic-poi-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString("hello")
	f.Close()

	data, err := source.Read(context.Background(), downloader.New(""), "file://"+f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("got %q", string(data))
	}
}

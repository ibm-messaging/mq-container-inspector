package container

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func TestNewContainerCopyConfig(t *testing.T) {
	execConfig := utils.ExecConfig{
		KubernetesConfig: &rest.Config{},
		Namespace:        "default",
		PodName:          "mypod",
		ContainerName:    "mycontainer",
	}
	srcDir := "/data"
	cfg := NewContainerCopyConfig(srcDir, execConfig)

	if cfg.KubeConfig != execConfig.KubernetesConfig {
		t.Errorf("KubernetesConfig not set correctly")
	}
	if cfg.Namespace != execConfig.Namespace {
		t.Errorf("Namespace not set correctly")
	}
	if cfg.PodName != execConfig.PodName {
		t.Errorf("PodName not set correctly")
	}
	if cfg.ContainerName != execConfig.ContainerName {
		t.Errorf("ContainerName not set correctly")
	}
	if cfg.SourcePath != srcDir {
		t.Errorf("SrcDir not set correctly")
	}
}

func TestCopyAll_Success(t *testing.T) {
	// Create a tar with a single file: /src/hello.txt
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	content := []byte("hello world")
	hdr := &tar.Header{
		Name: "/src/hello.txt",
		Mode: 0600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader failed: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("Close tar writer failed: %v", err)
	}

	destDir := t.TempDir()

	err := copyAll("/src", destDir, bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("copyAll failed: %v", err)
	}

	outFile := filepath.Join(destDir, "hello.txt")
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("Unexpected file content: got %q, want %q", string(data), "hello world")
	}
}

func TestCopyFileFromTar_Success(t *testing.T) {
	// Create a tar with a single file: /src/hello.txt
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	content := []byte("hello world")
	hdr := &tar.Header{
		Name: "/src/hello.txt",
		Mode: 0600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader failed: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("Close tar writer failed: %v", err)
	}

	pipeReader, pipeWriter := io.Pipe()
	cp := &copyPipe{
		copyConfig: CopyConfig{SourcePath: "/src/hello.txt"},
		reader:     pipeReader,
	}

	go func() {
		if _, err := pipeWriter.Write(buf.Bytes()); err != nil {
			pipeWriter.CloseWithError(err)
			return
		}
		pipeWriter.Close()
	}()

	outPath := filepath.Join(t.TempDir(), "output.txt")
	if err := copyAll("/src/hello.txt", outPath, cp); err != nil {
		t.Fatalf("copyAll for file failed: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Failed to read %q: %v", outPath, err)
	}

	if string(got) != string(content) {
		t.Errorf("Content mismatch: got %q, expected %q", got, content)
	}
}

func TestCopyAll_TarEntryOutsideDest(t *testing.T) {
	// Create a tar with an invalid path
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	content := []byte("bad stuff")
	hdr := &tar.Header{
		Name: "/src/../../evil.txt",
		Mode: 0600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader failed: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("Close tar writer failed: %v", err)
	}

	destDir := t.TempDir()

	err := copyAll("/src", destDir, bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatal("Expected error due to path escaping, got nil")
	}
	if !strings.Contains(err.Error(), "would write outside") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestNewCopyPipe_Initializes(t *testing.T) {
	cfg := CopyConfig{
		KubeConfig:    &rest.Config{},
		Namespace:     "ns",
		PodName:       "pod",
		ContainerName: "container",
		SourcePath:    "/src",
	}

	pipe := newCopyPipe(cfg, 2)

	if pipe == nil {
		t.Fatal("newCopyPipe returned nil")
	}
	if pipe.maxRetries != 2 {
		t.Errorf("maxRetries = %d; want 2", pipe.maxRetries)
	}
	if pipe.copyConfig != cfg {
		t.Errorf("copyConfig not set correctly")
	}
	if pipe.reader == nil || pipe.writer == nil {
		t.Error("reader or writer not initialized")
	}
}

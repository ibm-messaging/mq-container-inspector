package container

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// The configuration required to copy a folder from the source directory of a container.
type CopyConfig struct {
	KubeConfig    *rest.Config
	Namespace     string
	PodName       string
	ContainerName string
	SrcDir        string
}

// Helper function to create a CopyConfig from an ExecConfig
func NewContainerCopyConfig(srcDir string, execConfig utils.ExecConfig) CopyConfig {
	return CopyConfig{
		KubeConfig:    execConfig.KubernetesConfig,
		SrcDir:        srcDir,
		ContainerName: execConfig.ContainerName,
		PodName:       execConfig.PodName,
		Namespace:     execConfig.Namespace,
	}
}

// A pipe which connects a PipeWriter streaming the output of a tar command in the container
// with a PipeReader which copies the bytes into a directory.
type copyPipe struct {
	copyConfig CopyConfig
	maxRetries int
	reader     *io.PipeReader
	writer     *io.PipeWriter
	bytesRead  uint64
	retryCount int
}

// Create and start a copyPipe from a copyConfig and number of retries
func newCopyPipe(copyConfig CopyConfig, retries int) *copyPipe {
	t := &copyPipe{
		copyConfig: copyConfig,
		maxRetries: retries,
	}
	t.start(0)
	return t

}

// Start the copy pipe causing it to tar the src directory for the container
func (t *copyPipe) start(offset uint64) {
	t.reader, t.writer = io.Pipe()

	srcDir := strings.TrimSuffix(t.copyConfig.SrcDir, "/")
	baseCmd := []string{"tar", "cf", "-", srcDir}
	if t.maxRetries != 0 && offset > 0 {
		baseCmd = []string{"sh", "-c", fmt.Sprintf("tar cf - %s | tail -c+%d", srcDir, offset)}
	}

	copyCmd := utils.ExecConfig{
		KubernetesConfig: t.copyConfig.KubeConfig,
		PodName:          t.copyConfig.PodName,
		Namespace:        t.copyConfig.Namespace,
		ContainerName:    t.copyConfig.ContainerName,
		Cmd:              baseCmd,
	}

	exec, err := pods.ExecCmd(copyCmd)
	if err != nil {
		t.writer.CloseWithError(err)
		return
	}

	go func() {
		var stderr bytes.Buffer
		execErr := exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
			Stdout: t.writer,
			Stderr: &stderr,
		})

		if execErr != nil {
			t.writer.CloseWithError(fmt.Errorf("stream error: %w\nstderr: %s", execErr, stderr.String()))
		} else {
			t.writer.Close()
		}
	}()
}

// Read the data out of the copyPipe.
// Required to conform to the io.Reader interface
func (t *copyPipe) Read(p []byte) (int, error) {
	n, err := t.reader.Read(p)

	if err != nil {
		if t.maxRetries < 0 || t.retryCount < t.maxRetries {

			t.retryCount++

			// resume exactly after the last byte we delivered
			t.start(t.bytesRead + 1)
			err = nil
		} else {
			fmt.Printf("Dropping out copy after %d retries\n", t.retryCount)
		}
	} else {
		t.bytesRead += uint64(n)
	}
	return n, err
}

// Close the copyPipe reader
func (t *copyPipe) Close() error {
	return t.reader.Close()
}

// Copy all files from the srcDir to the destDir using a copyPipe
func copyAll(srcDir, destDir string, pipe io.Reader) error {
	pipeReader := tar.NewReader(pipe)
	cleanDest := filepath.Clean(destDir)

	for {
		h, err := pipeReader.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Validate prefix & build destination path
		if !strings.HasPrefix(h.Name, srcDir) {
			return fmt.Errorf("tar contents corrupted (entry %q lacks expected prefix %q)", h.Name, srcDir)
		}
		rel := strings.TrimPrefix(h.Name, srcDir)
		dstPath := filepath.Join(destDir, rel)
		cleanPath := filepath.Clean(dstPath)

		// Allow writing to destDir itself *or* to any child of it, but nowhere else.
		if cleanPath != cleanDest && !strings.HasPrefix(cleanPath, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("tar entry %q would write outside %q", h.Name, destDir)
		}

		if h.FileInfo().IsDir() {
			if err = os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			continue
		}

		// Ensure parent dir exists
		if err = os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		out, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		if _, err = io.Copy(out, pipeReader); err != nil {
			out.Close()
			return err
		}
		out.Close()
	}
}

// Copy the contents of a folder from the source directory and container specified in copyConfig
// to the outputDir with a fixed number of retries. Returns nil on success or an error on failure.
func CopyFolderToFile(copyConfig CopyConfig, outputDir string, retries int) error {

	copyPipe := newCopyPipe(copyConfig, retries)

	srcDir := strings.TrimLeft(strings.TrimSuffix(copyConfig.SrcDir, "/")+"/", "/")

	if err := copyAll(srcDir, outputDir, copyPipe); err != nil {
		return fmt.Errorf("error while copying runmqras files from container: %v", err)
	}

	return nil
}

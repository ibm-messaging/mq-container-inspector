/*
© Copyright IBM Corporation 2025

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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

// The configuration required to copy a the contents from the source path of a container.
type CopyConfig struct {
	KubeConfig    *rest.Config
	Namespace     string
	PodName       string
	ContainerName string
	SourcePath    string
}

// Helper function to create a CopyConfig from an ExecConfig
func NewContainerCopyConfig(srcPath string, execConfig utils.ExecConfig) CopyConfig {
	return CopyConfig{
		KubeConfig:    execConfig.KubernetesConfig,
		SourcePath:    srcPath,
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

// Start the copy pipe causing it to tar the src path for the container
func (t *copyPipe) start(offset uint64) {
	t.reader, t.writer = io.Pipe()

	srcPath := strings.TrimSuffix(t.copyConfig.SourcePath, "/")
	baseCmd := []string{"tar", "cf", "-", srcPath}
	if t.maxRetries != 0 && offset > 0 {
		baseCmd = []string{"sh", "-c", fmt.Sprintf("tar cf - %s | tail -c+%d", srcPath, offset)}
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

// Copy all files from the sourcePath to the destinationPath using a copyPipe
func copyAll(sourcePath, destinationPath string, pipe io.Reader) error {
	pipeReader := tar.NewReader(pipe)
	cleanDest := filepath.Clean(destinationPath)

	for {
		h, err := pipeReader.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Validate prefix & build destination path
		if !strings.HasPrefix(h.Name, sourcePath) {
			return fmt.Errorf("tar contents corrupted (entry %q lacks expected prefix %q)", h.Name, sourcePath)
		}
		rel := strings.TrimPrefix(h.Name, sourcePath)
		dstPath := filepath.Join(destinationPath, rel)
		cleanPath := filepath.Clean(dstPath)

		// Allow writing to destDir itself *or* to any child of it, but nowhere else.
		if cleanPath != cleanDest && !strings.HasPrefix(cleanPath, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("tar entry %q would write outside %q", h.Name, destinationPath)
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

// Copy the contents of the source path on the container specified in copyConfig
// to the outputPath with a fixed number of retries. Returns nil on success or an error on failure.
func CopyPathToFile(copyConfig CopyConfig, outputPath string, retries int) error {

	copyPipe := newCopyPipe(copyConfig, retries)

	sourcePath := copyConfig.SourcePath

	// if the sourcePath is a file then don't append a '/' at the end
	if filepath.Ext(sourcePath) != "" {
		sourcePath = strings.TrimLeft(strings.TrimSuffix(copyConfig.SourcePath, "/"), "/")
	} else {
		sourcePath = strings.TrimLeft(strings.TrimSuffix(copyConfig.SourcePath, "/")+"/", "/")
	}

	if err := copyAll(sourcePath, outputPath, copyPipe); err != nil {
		return fmt.Errorf("error while copying files from container: %v", err)
	}

	return nil
}

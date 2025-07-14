package tarzip

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func TarZipFolder(folderPath string) error {

	// prepare destination
	destination := folderPath + ".tar.gz"
	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}

	// setup gzip and tar
	gzipWriter := gzip.NewWriter(destinationFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// we dont want to tar the parent directory
	parentDir := filepath.Dir(folderPath)

	return filepath.WalkDir(folderPath, func(path string, dirEntry fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		info, err := dirEntry.Info()
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(parentDir, path)
		if err != nil {
			return err
		}

		// making sure that the path always have forward slashes '/'
		relativePath = strings.ReplaceAll(relativePath, string(os.PathSeparator), "/")

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = relativePath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil

	})

}

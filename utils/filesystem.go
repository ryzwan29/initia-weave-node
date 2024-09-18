package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// FileOrFolderExists checks if a file or folder exists at the given path.
func FileOrFolderExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DownloadAndExtractTarGz(url, tarballPath, extractedPath string) error {
	if err := DownloadFile(url, tarballPath); err != nil {
		return err
	}

	if err := ExtractTarGz(tarballPath, extractedPath); err != nil {
		return err
	}

	if err := os.Remove(tarballPath); err != nil {
		return fmt.Errorf("failed to remove tarball file: %v", err)
	}

	return nil
}

func DownloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func ExtractTarGz(src string, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.Create(target)
			if err != nil {
				return err
			}
			_, err = io.Copy(file, tarReader)
			file.Close()
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown type: %c", header.Typeflag)
		}
	}
	return nil
}

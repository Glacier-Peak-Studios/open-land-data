package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/jlaffaye/ftp"
)

// DownloadFile saves a file to the specified path
func DownloadFile(path string, urlStr string) (string, error) {
	err := os.MkdirAll(path, 0755)
	// Force unchecked certs
	if err != nil {
		log.Print(err)
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Fatal(err)
	}
	switch u.Scheme {
	case "http":
		return downloadHTTP(path, urlStr)
	case "https":
		return downloadHTTP(path, urlStr)
	case "ftp":
		return downloadFTP(path, u)
	default:
		return "", errors.New("URL type not supported")
	}
}

func downloadHTTP(path string, url string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := grab.NewClient()
	client.HTTPClient = &http.Client{Transport: tr}
	// Make Request
	req, _ := grab.NewRequest(path, url)
	resp := client.Do(req)
	if err := resp.Err(); err != nil {
		return "", err
	}
	fmt.Println("Download saved to", resp.Filename)
	return resp.Filename, nil
}

// "ftp://ftp.dfg.ca.gov/BDB/GIS/BIOS/Public_Datasets/1300_1399/ds1342.zip"
func downloadFTP(path string, u *url.URL) (string, error) {
	c, err := ftp.Dial(u.Hostname()+":21", ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return "", err
	}

	err = c.Login("anonymous", "anonymous")
	if err != nil {
		return "", err
	}

	// c.ChangeDir("desiredDir")

	res, err := c.Retr(u.Path)
	if err != nil {
		fmt.Println("Could not retrieve " + u.Path)
		return "", err
	}

	defer res.Close()

	outFile, err := os.Create(path + "/" + filepath.Base(u.Path))
	if err != nil {
		fmt.Println("Could not create dl file " + path + "/" + filepath.Base(u.Path))
		return "", err
	}

	defer outFile.Close()

	_, err = io.Copy(outFile, res)
	if err != nil {
		return "", err
	}

	return path + "/" + filepath.Base(u.Path), nil
}

func getFnameOnly(file string) string {
	var filename = filepath.Base(file)
	var extension = filepath.Ext(filename)
	return filename[0 : len(filename)-len(extension)]
}

// WalkMatch gets all files in root with specified pattern
func WalkMatch(root string, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

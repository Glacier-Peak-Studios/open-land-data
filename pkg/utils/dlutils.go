package utils

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/cavaliercoder/grab"
	"github.com/jlaffaye/ftp"
)

// DownloadFile saves a file to the specified path.
// Currently http, https, ftp, box, and file url types
// are supported
func DownloadFile(path string, urlStr string) (string, error) {
	// Ensure target path exists
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("Download type is %v", u.Scheme)
	switch u.Scheme {
	case "http":
		return downloadHTTP(path, urlStr)
	case "https":
		return downloadHTTP(path, urlStr)
	case "ftp":
		return downloadFTP(path, u)
	case "box":
		return downloadBox(path, urlStr)
	case "file":
		return downloadLocalFile(path, urlStr)
	default:
		return "", errors.New("URL type not supported")
	}
}

func downloadHTTP(path string, url string) (string, error) {
	// Force unchecked certs
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
	log.Debug().Msgf("Download saved to %v", resp.Filename)
	return resp.Filename, nil
}

func downloadFTP(path string, u *url.URL) (string, error) {
	c, err := ftp.Dial(u.Hostname()+":21", ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return "", err
	}
	err = c.Login("anonymous", "anonymous")
	if err != nil {
		return "", err
	}
	res, err := c.Retr(u.Path)
	if err != nil {
		log.Warn().Msg("Could not retrieve " + u.Path)
		return "", err
	}
	defer res.Close()
	outFile, err := os.Create(path + "/" + filepath.Base(u.Path))
	if err != nil {
		log.Warn().Msg("Could not create dl file " + path + "/" + filepath.Base(u.Path))
		return "", err
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, res)
	if err != nil {
		return "", err
	}
	return path + "/" + filepath.Base(u.Path), nil
}

func downloadBox(path, boxpath string) (string, error) {
	boxdir, exists := os.LookupEnv("BOXPATH")
	if !exists {
		return "", errors.New("failed to fetch box source: BOXPATH env not set")
	}
	log.Debug().Msgf("BOXPATH=%v", boxdir)
	pathIn := strings.Replace(boxpath, "box://", "file://"+boxdir+"/", 1)
	return downloadLocalFile(path, pathIn)
}

func downloadLocalFile(pathOut, pathIn string) (string, error) {
	log.Debug().Msgf("local file is: %v", pathIn)
	pathIn = strings.Replace(pathIn, "file://", "", 1)
	_, err := RunCommandDefault("cp", "-r", pathIn, pathOut)
	if err != nil {
		log.Warn().Msgf("Could not copy local file: %v to %v", pathIn, pathOut)
		return "", err
	}
	return pathOut + "/" + filepath.Base(pathIn), nil
}

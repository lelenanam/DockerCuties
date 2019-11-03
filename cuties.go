package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/google/go-github/github"
	"github.com/lelenanam/downsize"
	"github.com/lelenanam/screenshot"
	log "github.com/sirupsen/logrus"
)

//StartCutiePullReq is the number of pull request
//where new template (with a picture of a cute animal) was added
const StartCutiePullReq = 20514

//Repo is repository with docker cuties
const Repo = "moby"

//Owner of repository
const Owner = "moby"

//SearchIssuesLimit is github limit for Search.Issues results
const SearchIssuesLimit = 1000

//PerPage number of pull requests per page
const PerPage = 50

//TwitterUser is account for docker cuties in twitter
const TwitterUser = "DockerCuties"

//ProjectOwner is a twitter account for notifications
const ProjectOwner = "lelenanam"

//TwitterUploadLimit is limit for media upload in bytes
const TwitterUploadLimit = 3145728

// GetURLFromPull parses the body of the pull request and return last image URL if found
// try to find flickr first
func GetURLFromPull(pull *github.Issue) string {
	str := *pull.Body
	res := ""

	// Example:
	// [![kitteh](https://c2.staticflickr.com/4/3147/2567501805_17ee8fd947_z.jpg)](https://flic.kr/p/4UT7Qv)
	flickre := regexp.MustCompile(`\[!\[.*\]\((.*)\)\]\(.*\)`)
	flickResult := flickre.FindAllStringSubmatch(str, -1)

	if len(flickResult) > 0 {
		lastres := flickResult[len(flickResult)-1]
		if len(lastres) > 1 {
			res := lastres[len(lastres)-1]
			res = strings.SplitN(res, " ", 2)[0]
			return res
		}
	}

	// Example:
	// ![image](https://cloud.githubusercontent.com/assets/2367858/23283487/02bb756e-f9db-11e6-9aa8-5f3e1bb80df3.png  "Swans")
	imagere := regexp.MustCompile(`!\[.*\]\((.*)\)`)
	imageResult := imagere.FindAllStringSubmatch(str, -1)

	if len(imageResult) > 0 {
		lastres := imageResult[len(imageResult)-1]
		if len(lastres) > 1 {
			res := lastres[len(lastres)-1]
			res = strings.SplitN(res, " ", 2)[0]
			return res
		}
	}
	return res
}

// GetImageFromURL downloads an image from url and returns image img, its size, format and error
// for animated gif format returns data in gifByte slice
func GetImageFromURL(url string) (img image.Image, format string, size int, gifByte []byte, err error) {
	log.WithFields(log.Fields{"URL": url}).Debug("Download")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(log.Fields{"URL": url}).WithError(err).Error("Cannot create new GET request")
		return nil, "", 0, nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 Firefox/26.0")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"URL": url}).WithError(err).Error("Cannot download")
		return nil, "", 0, nil, err
	}
	log.WithFields(log.Fields{"Content Length": res.ContentLength, "Status Code": res.StatusCode}).Debug("Got")

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.WithFields(log.Fields{"URL": url}).WithError(err).Error("Cannot close body")
		}
	}()

	var reader io.Reader = res.Body

	if res.Header["Content-Type"][0] == "image/gif" {
		gifByte, err = ioutil.ReadAll(res.Body)
		if err != nil {
			log.WithFields(log.Fields{"URL": url}).WithError(err).Error("Cannot read body")
			return nil, "gif", 0, nil, err
		}
		reader = bytes.NewReader(gifByte)
	}

	img, format, err = image.Decode(reader)
	if err != nil {
		log.WithFields(log.Fields{"URL": url}).WithError(err).Error("Cannot decode image")
		return nil, "", 0, nil, err
	}

	log.WithFields(log.Fields{"format": format, "size": int(res.ContentLength)}).Debug("Image decoded")
	return img, format, int(res.ContentLength), gifByte, nil
}

// GetStringFromImage returns string of cutie image img or error
func GetStringFromImage(img image.Image, format string, size int, gifByte []byte) (string, error) {
	b := bytes.NewBuffer(nil)
	encoder := base64.NewEncoder(base64.StdEncoding, b)

	// Need to resize image
	if size >= TwitterUploadLimit || size < 0 {
		log.WithFields(log.Fields{"Twitter upload limit": TwitterUploadLimit}).Debug("Downsize image")
		opts := &downsize.Options{Size: TwitterUploadLimit, Format: format}
		err := downsize.Encode(encoder, img, opts)
		if err != nil {
			log.WithFields(log.Fields{"format": format}).WithError(err).Error("Cannot downsize image")
			return "", err
		}
	} else {
		if format == "gif" {
			gifReader := bytes.NewReader(gifByte)
			_, err := io.Copy(encoder, gifReader)
			if err != nil {
				log.WithFields(log.Fields{"format": format}).WithError(err).Error("Cannot copy gif image")
				return "", err
			}
		} else {
			// no need to resize, just encode
			if err := ImageEncode(encoder, img, format); err != nil {
				log.WithFields(log.Fields{"format": format}).WithError(err).Error("Cannot encode image")
				return "", err
			}
		}
	}

	if err := encoder.Close(); err != nil {
		log.WithError(err).Error("Cannot close encoder")
		return "", err
	}

	if b.Len() == 0 {
		log.Warn("Empty image data")
		return "", nil
	}
	return b.String(), nil
}

var errIsScreenshot = errors.New("picture is screenshot")
var errImageNotFound = errors.New("picture not found")

// GetCutieFromPull returns string of cutie image from pull request pull
func GetCutieFromPull(pull *github.Issue, screenCheck bool) (string, error) {
	url := GetURLFromPull(pull)
	if url == "" {
		return "", errImageNotFound
	}
	img, format, size, gifByte, err := GetImageFromURL(url)
	if err != nil {
		return "", errors.Wrap(err, "cannot get image from URL")
	}

	if screenCheck && screenshot.Detect(img) {
		return "", errIsScreenshot
	}

	str, err := GetStringFromImage(img, format, size, gifByte)
	if err != nil {
		return "", errors.Wrap(err, "cannot get string for image")
	}
	return str, nil
}

// ImageEncode encodes image m with format to writer w
func ImageEncode(w io.Writer, m image.Image, format string) error {
	switch format {
	case "jpeg":
		return jpeg.Encode(w, m, &jpeg.Options{Quality: 95})
	case "png":
		return png.Encode(w, m)
	case "gif":
		return gif.Encode(w, m, nil)
	default:
		return fmt.Errorf("Unknown image format %q", format)
	}
}

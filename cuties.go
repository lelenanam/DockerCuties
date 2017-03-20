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
	"net/http"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/lelenanam/downsize"
	"github.com/lelenanam/screenshot"
)

//StartCutiePullReq is the number of pull request
//where new template (with a picture of a cute animal) was added
const StartCutiePullReq = 20514

//Repo is repository with docker cuties
const Repo = "docker"

//Owner of repository
const Owner = "docker"

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

// GetURLFromPull parses the body of the pull request and return an image URL if found
func GetURLFromPull(pull *github.Issue) string {
	if strings.Contains(*pull.Body, "flickr.com") {
		log.WithFields(log.Fields{"body": *pull.Body, "pull": *pull.URL}).Warn("flic.kr found")
		// [![kitteh](https://c2.staticflickr.com/4/3147/2567501805_17ee8fd947_z.jpg)](https://flic.kr/p/4UT7Qv)
		re := regexp.MustCompile(`\[!\[.*\]\((.*)\)\]\(.*\)`)
		result := re.FindStringSubmatch(*pull.Body)
		if len(result) > 1 {
			return result[len(result)-1]
		}
		return ""
	}

	// ![image](https://cloud.githubusercontent.com/assets/2367858/23283487/02bb756e-f9db-11e6-9aa8-5f3e1bb80df3.png)
	re := regexp.MustCompile(`!\[.*\]\((.*)\)`)
	result := re.FindStringSubmatch(*pull.Body)
	if len(result) > 1 {
		return result[len(result)-1]
	}
	return ""
}

// GetImageFromURL downloads an image from url and returns image img, its size, format and error
func GetImageFromURL(url string) (img image.Image, format string, size int, err error) {
	log.WithFields(log.Fields{"URL": url}).Debug("Download")
	res, err := http.Get(url)
	if err != nil {
		log.WithFields(log.Fields{"URL": url}).WithError(err).Error("Cannot download")
		return nil, "", 0, err
	}
	log.WithFields(log.Fields{"Content Length": res.ContentLength, "Status Code": res.StatusCode}).Debug("Got")

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.WithFields(log.Fields{"URL": url}).WithError(err).Error("Cannot close body")
		}
	}()

	img, format, err = image.Decode(res.Body)
	log.WithFields(log.Fields{"format": format, "size": int(res.ContentLength)}).Debug("Image decoded")
	if err != nil {
		log.WithFields(log.Fields{"URL": url}).WithError(err).Error("Cannot decode image")
		return nil, "", 0, err
	}
	return img, format, int(res.ContentLength), nil
}

// GetStringFromImage returns string of cutie image img or error
func GetStringFromImage(img image.Image, format string, size int) (string, error) {
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
	} else { // no need to resize, just encode
		if err := ImageEncode(encoder, img, format); err != nil {
			log.WithFields(log.Fields{"format": format}).WithError(err).Error("Cannot encode image")
			return "", err
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

// GetCutieFromPull returns string of cutie image from pull request pull
func GetCutieFromPull(pull *github.Issue) string {
	url := GetURLFromPull(pull)
	if url != "" {
		img, format, size, err := GetImageFromURL(url)
		if err != nil {
			log.WithFields(log.Fields{"URL": url}).WithError(err).Error("Cannot get image from URL")
			return ""
		}
		if screenshot.Detect(img) {
			return "screenshot"
		}
		str, err := GetStringFromImage(img, format, size)
		if err != nil {
			log.WithFields(log.Fields{"Pull request": pull.Number}).WithError(err).Error("Cannot get string for image")
			return ""
		}
		return str
	}
	return ""
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

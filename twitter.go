package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/ChimeraCoder/anaconda"
	"github.com/lelenanam/downsize"
)

// TwitterTokens provides credentials for accessing twitter
type TwitterTokens struct {
	twitterConsumerKey    string
	twitterConsumerSecret string
	twitterAccessToken    string
	twitterAccessSecret   string
}

// Twitter provides api for twitter
type Twitter struct {
	api *anaconda.TwitterApi
}

// NewTwitter returns *Twitter to work with twitter
func NewTwitter(t TwitterTokens) *Twitter {
	anaconda.SetConsumerKey(t.twitterConsumerKey)
	anaconda.SetConsumerSecret(t.twitterConsumerSecret)
	api := anaconda.NewTwitterApi(t.twitterAccessToken, t.twitterAccessSecret)
	return &Twitter{api: api}
}

// DeleteAllTweets delets all tweets for user
func (t *Twitter) DeleteAllTweets(user string) error {
	api := t.api

	v := url.Values{}
	v.Set("user", user)

	timeline, err := api.GetUserTimeline(v)
	if err != nil {
		log.WithFields(log.Fields{"values": v}).WithError(err).Error("Cannot get user timeline")
		return err
	}

	if len(timeline) == 0 {
		return nil
	}

	for _, tw := range timeline {
		log.WithFields(log.Fields{"ID": tw.Id, "Text": tw.Text}).Info("Delete tweet")
		_, err := api.DeleteTweet(tw.Id, false)
		if err != nil {
			log.WithFields(log.Fields{"ID": tw.Id, "Text": tw.Text}).WithError(err).Error("Cannot delete tweet")
			continue
		}
	}

	oldest := timeline[len(timeline)-1].Id - 1
	for len(timeline) > 0 {
		v.Set("max_id", strconv.Itoa(int(oldest)))
		timeline, err := api.GetUserTimeline(v)
		if err != nil {
			return err
		}

		for _, tw := range timeline {
			log.WithFields(log.Fields{"ID": tw.Id, "Text": tw.Text}).Info("Delete tweet")
			_, err := api.DeleteTweet(tw.Id, false)
			if err != nil {
				log.WithFields(log.Fields{"ID": tw.Id, "Text": tw.Text}).WithError(err).Error("Cannot delete tweet")
				continue
			}
		}
		if len(timeline) == 0 {
			return nil
		}
		oldest = timeline[len(timeline)-1].Id - 1
	}
	return nil
}

// LastPostedPull returns last number of pull request posted to twitter
// if pull request link not found in twitter timeline, returns -1
func (t *Twitter) LastPostedPull() (int, error) {
	api := t.api

	v := url.Values{}
	v.Set("user", TwitterUser)

	timeline, err := api.GetUserTimeline(v)
	if err != nil {
		return -1, err
	}

	for _, tw := range timeline {
		log.WithFields(log.Fields{"Created At": tw.CreatedAt, "Text": tw.Text}).Info("Twitter timeline")
	}

	if len(timeline) == 0 {
		return -1, nil
	}

	withURLs := -1
	for i, tw := range timeline {
		if len(tw.Entities.Urls) > 0 {
			withURLs = i
			break
		}
	}
	if withURLs == -1 {
		return -1, nil
	}

	urls := timeline[withURLs].Entities.Urls
	var lastpull string
	for _, u := range urls {
		log.WithFields(log.Fields{"URL": u.Expanded_url}).Info("Last posted")
		lastpull = u.Expanded_url //last url in tweet
	}
	splited := strings.Split(lastpull, "/")
	lastNstr := splited[len(splited)-1]
	if n, err := strconv.Atoi(lastNstr); err == nil {
		return n, nil
	}
	return -1, nil
}

// PostToTwitter posts cutie to twitter
func (t *Twitter) PostToTwitter(cutie *DockerCutie) error {
	api := t.api

	log.WithFields(log.Fields{"URL": cutie.cutieURL}).Info("Download")
	res, err := http.Get(cutie.cutieURL)
	if err != nil {
		log.WithFields(log.Fields{"URL": cutie.cutieURL}).WithError(err).Error("Cannot download")
		return nil
	}

	log.WithFields(log.Fields{"Content Length": res.ContentLength, "Status Code": res.StatusCode}).Info("Got")
	defer res.Body.Close()

	b := bytes.NewBuffer(nil)
	encoder := base64.NewEncoder(base64.StdEncoding, b)
	defer encoder.Close()

	if res.ContentLength >= TwitterUploadLimit || res.ContentLength < 0 {
		img, format, err := image.Decode(res.Body)
		if err != nil {
			log.WithFields(log.Fields{"Body": res.Body}).WithError(err).Error("Cannot decode image")
			return nil
		}
		log.WithFields(log.Fields{"Twitter upload limit": TwitterUploadLimit}).Info("Downsize image")
		opts := &downsize.Options{Size: TwitterUploadLimit, Format: format}
		err = downsize.Encode(encoder, img, opts)
		if err != nil {
			log.WithFields(log.Fields{"Body": res.Body}).WithError(err).Error("Cannot downsize")
			return nil
		}
	} else {
		_, err := io.Copy(encoder, res.Body)
		if err != nil {
			log.WithFields(log.Fields{"Body": res.Body}).WithError(err).Error("Cannot copy to writer")
			return nil
		}
	}
	encoder.Close()

	if b.Len() == 0 {
		log.WithFields(log.Fields{"Body": res.Body}).Warn("Empty image data")
		return nil
	}

	mediaResponse, err := api.UploadMedia(b.String())
	if err != nil {
		log.WithFields(log.Fields{"String of data": b.String()}).WithError(err).Error("Cannot upload data")
		return nil
	}
	log.WithFields(log.Fields{"MediaID": mediaResponse.MediaID}).Info("Uploaded")

	v := url.Values{}
	v.Set("media_ids", strconv.FormatInt(mediaResponse.MediaID, 10))
	msg := cutie.pullURL
	_, err = api.PostTweet(msg, v)
	if err != nil {
		log.WithFields(log.Fields{"Tweet message": msg}).WithError(err).Error("Cannot post tweet")
		return nil
	}
	return nil
}

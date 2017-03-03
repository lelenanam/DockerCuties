package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/morelena/downsize"
)

// Twitter provides credentials for accessing twitter
type Twitter struct {
	twitterConsumerKey    string
	twitterConsumerSecret string
	twitterAccessToken    string
	twitterAccessSecret   string
}

// DeleteAllTweets delets all tweets for user
func (t Twitter) DeleteAllTweets(user string) error {
	anaconda.SetConsumerKey(t.twitterConsumerKey)
	anaconda.SetConsumerSecret(t.twitterConsumerSecret)
	api := anaconda.NewTwitterApi(t.twitterAccessToken, t.twitterAccessSecret)

	v := url.Values{}
	v.Set("user", user)

	timeline, err := api.GetUserTimeline(v)
	if err != nil {
		return err
	}

	if len(timeline) == 0 {
		return nil
	}

	for _, tw := range timeline {
		log.Println("Delete tweet:", tw.Id, tw.Text)
		_, err := api.DeleteTweet(tw.Id, false)
		if err != nil {
			log.Println("Cannot delete tweet:", err)
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
			log.Println("Delete tweet:", tw.Id, tw.Text)
			_, err := api.DeleteTweet(tw.Id, false)
			if err != nil {
				log.Println("Cannot delete tweet:", err)
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
func (t Twitter) LastPostedPull() (int, error) {
	anaconda.SetConsumerKey(t.twitterConsumerKey)
	anaconda.SetConsumerSecret(t.twitterConsumerSecret)
	api := anaconda.NewTwitterApi(t.twitterAccessToken, t.twitterAccessSecret)

	v := url.Values{}
	v.Set("user", TwitterUser)

	timeline, err := api.GetUserTimeline(v)
	if err != nil {
		return -1, err
	}

	for _, tw := range timeline {
		log.Println("Twitter timeline:\n", tw.CreatedAt, tw.Entities.Urls)
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
		log.Println("Last posted URL:", u.Expanded_url)
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
func (t Twitter) PostToTwitter(cutie *DockerCuties) error {
	anaconda.SetConsumerKey(t.twitterConsumerKey)
	anaconda.SetConsumerSecret(t.twitterConsumerSecret)
	api := anaconda.NewTwitterApi(t.twitterAccessToken, t.twitterAccessSecret)

	log.Println("Download from:", cutie.cutieURL)
	res, err := http.Get(cutie.cutieURL)
	if err != nil {
		return fmt.Errorf("Cannot download cutie: %q", err)
	}

	log.Println("Got:", res.ContentLength, "StatusCode:", res.StatusCode)
	defer res.Body.Close()

	if res.ContentLength <= 0 {
		return fmt.Errorf("Response length <= 0")
	}

	b := bytes.NewBuffer(nil)
	encoder := base64.NewEncoder(base64.StdEncoding, b)
	defer encoder.Close()
	if res.ContentLength >= TwitterUploadLimit {
		log.Println("Downsize image to twitter limit:", TwitterUploadLimit)
		err = downsize.Downsize(TwitterUploadLimit, res.Body, encoder)
		if err != nil {
			return fmt.Errorf("Cannot downsize: %q", err)
		}
	} else {
		_, err := io.Copy(encoder, res.Body)
		if err != nil {
			return err
		}
	}
	encoder.Close()
	mediaResponse, err := api.UploadMedia(b.String())
	if err != nil {
		return fmt.Errorf("Cannot upload data: %q", err)
	}
	log.Println("Uploaded, mediaID:", mediaResponse.MediaID)

	v := url.Values{}
	v.Set("media_ids", strconv.FormatInt(mediaResponse.MediaID, 10))
	_, err = api.PostTweet(cutie.pullURL, v)
	if err != nil {
		return fmt.Errorf("Cannot post tweet: %q", err)
	}
	return nil
}

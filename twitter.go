package main

import (
	"net/url"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/ChimeraCoder/anaconda"
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
		log.WithFields(log.Fields{"ID": tw.Id, "Text": tw.Text}).Debug("Delete tweet")
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
			log.WithFields(log.Fields{"ID": tw.Id, "Text": tw.Text}).Debug("Delete tweet")
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
		log.WithFields(log.Fields{"Created At": tw.CreatedAt, "Text": tw.Text}).Debug("Twitter timeline")
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
		log.WithFields(log.Fields{"URL": u.Expanded_url}).Debug("Last posted")
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
func (t *Twitter) PostToTwitter(cutie string, msg string) error {
	api := t.api
	mediaResponse, err := api.UploadMedia(cutie)
	if err != nil {
		log.WithFields(log.Fields{"String of data": cutie}).WithError(err).Error("Cannot upload data")
		return err
	}
	log.WithFields(log.Fields{"MediaID": mediaResponse.MediaID}).Debug("Uploaded")

	v := url.Values{}
	v.Set("media_ids", strconv.FormatInt(mediaResponse.MediaID, 10))
	_, err = api.PostTweet(msg, v)
	if err != nil {
		log.WithFields(log.Fields{"Tweet message": msg}).WithError(err).Error("Cannot post tweet")
		return err
	}
	return nil
}

// Notify notifies ProjectOwner about screenshot and errors with direct message in twitter
func (t *Twitter) Notify(msg string) {
	api := t.api
	_, err := api.PostDMToScreenName(msg, ProjectOwner)
	if err != nil {
		log.WithFields(log.Fields{"Message": msg}).WithError(err).Error("Cannot notify")
	}
}

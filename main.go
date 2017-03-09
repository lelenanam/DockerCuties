package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
)

func main() {
	tokens, err := LoadTokens()
	twitter := NewTwitter(tokens.twitter)
	if err != nil {
		log.Println(err)
		return
	}

	if err := twitter.DeleteAllTweets(TwitterUser); err != nil {
		log.WithFields(log.Fields{"User": TwitterUser}).WithError(err).Error("Cannot delete all tweets")
		return
	}

	// tweetCutie posts cutie from pull request pull to twitter
	tweetCutie := func(pull *github.Issue) error {
		if pull.Body != nil {
			cutie := GetCutieFromPull(pull)
			if cutie != nil {
				log.WithFields(log.Fields{"number": cutie.pullnumber, "pull": cutie.pullURL, "cutie": cutie.cutieURL}).Info("Cutie")
				if err := twitter.PostToTwitter(cutie); err != nil {
					return err
				}
			}
		}
		return nil
	}

	lastPosted, err := twitter.LastPostedPull()
	if err != nil {
		log.WithError(err).Error("Cannot check last posted pull request")
		return
	}
	if lastPosted > 0 {
		if err = tokens.github.PullsSinceFunc(lastPosted+1, tweetCutie); err != nil {
			log.WithFields(log.Fields{"since": lastPosted}).WithError(err).Error("For pull requests since")
			return
		}
	} else {
		if err = tokens.github.PullsSinceFunc(StartCutiePullReq, tweetCutie); err != nil {
			log.WithFields(log.Fields{"since": StartCutiePullReq}).WithError(err).Error("For pull requests since")
			return
		}
	}
}

package main

import (
	"flag"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
)

var isDelete = flag.Bool("delete", false, "delete all tweets before posting")

func updateTwitter(g *Github, t *Twitter) {
	// tweetCutie posts cutie from pull request pull to twitter
	tweetCutie := func(pull *github.Issue) error {
		if pull.Body != nil {
			cutie := GetCutieFromPull(pull)
			if cutie != nil {
				log.WithFields(log.Fields{"number": cutie.pullnumber, "pull": cutie.pullURL, "cutie": cutie.cutieURL}).Info("Cutie")
				if err := t.PostToTwitter(cutie); err != nil {
					return err
				}
			}
		}
		return nil
	}

	lastPosted, err := t.LastPostedPull()
	if err != nil {
		log.WithError(err).Error("Cannot check last posted pull request")
		return
	}
	if lastPosted > 0 {
		if err = g.PullsSinceFunc(lastPosted+1, tweetCutie); err != nil {
			log.WithFields(log.Fields{"since": lastPosted + 1}).WithError(err).Error("For pull requests since")
			return
		}
	} else {
		if err = g.PullsSinceFunc(StartCutiePullReq, tweetCutie); err != nil {
			log.WithFields(log.Fields{"since": StartCutiePullReq}).WithError(err).Error("For pull requests since")
			return
		}
	}
}

func main() {
	flag.Parse()

	tokens, err := LoadTokens()
	twitter := NewTwitter(tokens.twitter)
	gh := NewGithub(tokens.github)
	if err != nil {
		log.Println(err)
		return
	}

	if *isDelete {
		if err := twitter.DeleteAllTweets(TwitterUser); err != nil {
			log.WithFields(log.Fields{"User": TwitterUser}).WithError(err).Error("Cannot delete all tweets")
			return
		}
	}

	// Single post by number
	// n := 31705
	// if err = tokens.github.PullFunc(n, tweetCutie); err != nil {
	// 	log.WithFields(log.Fields{"number": n}).WithError(err).Error("For pull request")
	// 	return
	// }
	// return

	for range time.Tick(60 * time.Second) {
		updateTwitter(gh, twitter)
	}
}

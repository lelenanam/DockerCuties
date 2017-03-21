package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
)

var isDelete = flag.Bool("delete", false, "delete all tweets before posting")
var logLevel = flag.String("loglevel", "warning", "log level (panic, fatal, error, warn or warning, info, debug)")
var lastPosted int

func updateTwitter(g *Github, t *Twitter) {
	// tweetCutie posts cutie from pull request pull to twitter
	tweetCutie := func(pull *github.Issue) error {
		if pull.Body != nil {
			cutie := GetCutieFromPull(pull)
			if cutie == "screenshot" {
				log.WithFields(log.Fields{"number": *pull.Number, "URL": *pull.HTMLURL}).Warn("Screenshot detected")
				t.Notify(fmt.Sprintf("Screenshot detected: %s", *pull.HTMLURL))
				return nil
			}
			if cutie != "" {
				log.WithFields(log.Fields{"number": *pull.Number, "URL": *pull.HTMLURL}).Info("Cutie")
				msg := fmt.Sprintf("%s #dockercuties #docker", *pull.HTMLURL)
				if err := t.PostToTwitter(cutie, msg); err != nil {
					return err
				}
				lastPosted = *pull.Number
			}
		}
		return nil
	}

	if lastPosted > 0 {
		if err := g.PullsSinceFunc(lastPosted+1, tweetCutie); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				log.WithFields(log.Fields{"Owner": Owner, "Repo": Repo, "number": lastPosted + 1}).Debug("Issue not found")
				return
			}
			log.WithFields(log.Fields{"since": lastPosted + 1}).WithError(err).Error("For pull requests since")
			return
		}
	} else {
		if err := g.PullsSinceFunc(StartCutiePullReq, tweetCutie); err != nil {
			log.WithFields(log.Fields{"since": StartCutiePullReq}).WithError(err).Error("For pull requests since")
			return
		}
	}
}

func main() {
	flag.Parse()
	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.WithFields(log.Fields{"log level": *logLevel}).WithError(err).Fatal("Cannot parse log level")
	}
	log.SetLevel(lvl)

	tokens, err := LoadTokens()
	log.Info("Tokens are loaded")
	if err != nil {
		log.WithError(err).Fatal("Cannot parse tokens")
		return
	}

	twitter := NewTwitter(tokens.twitter)
	log.Info("Connect to twitter")
	gh := NewGithub(tokens.github)
	log.Info("Connect to github")

	if *isDelete {
		if err := twitter.DeleteAllTweets(TwitterUser); err != nil {
			log.WithFields(log.Fields{"User": TwitterUser}).WithError(err).Error("Cannot delete all tweets")
			return
		}
	}

	// tweetCutie := func(pull *github.Issue) error {
	// 	if pull.Body != nil {
	// 		msg := fmt.Sprintf("%s #dockercuties #docker", *pull.HTMLURL)
	// 		cutie := GetCutieFromPull(pull)
	// 		if cutie == "screenshot" {
	// 			log.WithFields(log.Fields{"number": *pull.Number, "RUL": *pull.HTMLURL}).Warn("Screenshot detected")
	// 			twitter.Notify(msg)
	// 			return nil
	// 		}
	// 		if cutie != "" {
	// 			log.WithFields(log.Fields{"number": *pull.Number}).Info("Cutie")
	// 			if err := twitter.PostToTwitter(cutie, msg); err != nil {
	// 				return err
	// 			}
	// 			lastPosted = *pull.Number
	// 		}
	// 	}
	// 	return nil
	// }
	// Single post by number
	// n := 31933
	// if err = gh.PullFunc(n, tweetCutie); err != nil {
	// 	log.WithFields(log.Fields{"number": n}).WithError(err).Error("For pull request")
	// 	return
	// }
	// return

	lastPosted = twitter.LastPostedPull()

	for range time.Tick(60 * time.Second) {
		updateTwitter(gh, twitter)
	}
}

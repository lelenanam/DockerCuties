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
		if pull.Body == nil {
			return nil
		}
		cutie, err := GetCutieFromPull(pull)
		if err != nil {
			switch err {
			case errImageNotFound:

			case errIsScreenshot:
				log.WithFields(log.Fields{"number": *pull.Number, "URL": *pull.HTMLURL}).Warn("Screenshot detected")
				t.Notify(fmt.Sprintf("Screenshot detected: %s", *pull.HTMLURL))
			default:
				log.WithFields(log.Fields{"since": lastPosted + 1}).WithError(err).Error("For pull requests since")
				t.Notify(fmt.Sprintf("Cannot get cutie from pull request %d, %s: %s", *pull.Number, *pull.HTMLURL, err))
				// return err
			}
			return nil
		}
		if cutie != "" {
			log.WithFields(log.Fields{"number": *pull.Number, "URL": *pull.HTMLURL}).Info("Cutie")
			msg := fmt.Sprintf("%s #dockercuties #docker", *pull.HTMLURL)
			if err := t.PostToTwitter(cutie, msg); err != nil {
				t.Notify(fmt.Sprintf("Cannot post tweet: %s", err))
				return err
			}
			lastPosted = *pull.Number
		}
		return nil
	}
	log.WithFields(log.Fields{"number": lastPosted}).Debug("Last posted")
	if lastPosted > 0 {
		if err := g.PullsSinceFunc(lastPosted+1, tweetCutie); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				log.WithFields(log.Fields{"Owner": Owner, "Repo": Repo, "number": lastPosted + 1}).Debug("Issue not found")
				return
			}
			log.WithFields(log.Fields{"since": lastPosted + 1}).WithError(err).Error("For pull requests since")
			t.Notify(fmt.Sprintf("Error for pull requests since %d: %s", lastPosted+1, err))
			return
		}
	} else {
		if err := g.PullsSinceFunc(StartCutiePullReq, tweetCutie); err != nil {
			log.WithFields(log.Fields{"since": StartCutiePullReq}).WithError(err).Error("For pull requests since")
			t.Notify(fmt.Sprintf("Error for pull requests since %d: %s", StartCutiePullReq, err))
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

	t := NewTwitter(tokens.twitter)
	log.Info("Connect to twitter")
	gh := NewGithub(tokens.github)
	log.Info("Connect to github")

	if *isDelete {
		if err := t.DeleteAllTweets(TwitterUser); err != nil {
			log.WithFields(log.Fields{"User": TwitterUser}).WithError(err).Error("Cannot delete all tweets")
			t.Notify(fmt.Sprintf("Cannot delete all tweets: %s", err))
			return
		}
	}

	// // Single post by number
	// n := 32085
	// if err = gh.PullFunc(n, tweetCutie); err != nil {
	// 	log.WithFields(log.Fields{"number": n}).WithError(err).Error("For pull request")
	// 	return
	// }
	// return

	lastPosted = t.LastPostedPull()

	for range time.Tick(60 * time.Second) {
		updateTwitter(gh, t)
	}
}

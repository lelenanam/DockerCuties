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

// Delay for update in seconds
const defaultDelay = 60 * time.Second

// Number of attempts to post if error occurred
const maxAttempts = 3

// Number of current attempt
var attempt = 0

func updateTwitter(g *Github, t *Twitter) error {
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
				t.Notify(fmt.Sprintf("%d Screenshot detected: %s", *pull.Number, *pull.HTMLURL))
				lastPosted = *pull.Number
			default:
				log.WithFields(log.Fields{"since": lastPosted + 1, "attempt": attempt, "PullNumber": *pull.Number}).WithError(err).Error("For pull requests since")
				attempt++
				if attempt == maxAttempts {
					t.Notify(fmt.Sprintf("Cannot get cutie from pull request %d, %s: %s", *pull.Number, *pull.HTMLURL, err))
					lastPosted = *pull.Number
				}
				return err
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
				return err
			}
			log.WithFields(log.Fields{"since": lastPosted + 1}).WithError(err).Error("For pull requests since")
			t.Notify(fmt.Sprintf("Error for pull requests since %d: %s", lastPosted+1, err))
			return err
		}
	} else {
		if err := g.PullsSinceFunc(StartCutiePullReq, tweetCutie); err != nil {
			log.WithFields(log.Fields{"since": StartCutiePullReq}).WithError(err).Error("For pull requests since")
			t.Notify(fmt.Sprintf("Error for pull requests since %d: %s", StartCutiePullReq, err))
			return err
		}
	}
	return nil
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
	// n := 21825
	// if err = gh.PullFunc(n, tweetCutie); err != nil {
	// 	log.WithFields(log.Fields{"number": n}).WithError(err).Error("For pull request")
	// 	return
	// }
	// return

	lastPosted = t.LastPostedPull()

	delay := defaultDelay

	for {
		if err := updateTwitter(gh, t); err != nil {
			if delay < time.Duration(1800*time.Second) {
				delay *= 2
			}
		} else {
			delay = defaultDelay
		}
		time.Sleep(delay)
	}
}

package main

import (
	"log"

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
		log.Println(err)
		return
	}

	// tweetCutie posts cutie from pull request pull to twitter
	tweetCutie := func(pull *github.Issue) error {
		if pull.Body != nil {
			cutie := GetCutieFromPull(pull)
			if cutie != nil {
				log.Println("Cutie:", *cutie)
				if err := twitter.PostToTwitter(cutie); err != nil {
					return err
				}
			}
		}
		return nil
	}

	lastPosted, err := twitter.LastPostedPull()
	if err != nil {
		log.Println(err)
		return
	}
	if lastPosted > 0 {
		tokens.github.PullsSinceFunc(lastPosted+1, tweetCutie)
	} else {
		tokens.github.PullsSinceFunc(StartCutiePullReq, tweetCutie)
	}
}

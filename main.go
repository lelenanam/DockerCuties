package main

import (
	"log"

	"github.com/google/go-github/github"
)

func main() {
	tokens, err := LoadTokens()
	if err != nil {
		log.Println(err)
		return
	}

	if err := tokens.twitter.DeleteAllTweets(TwitterUser); err != nil {
		log.Println(err)
		return
	}

	// TweetCutie posts cutie from pull request pull to twitter
	TweetCutie := func(pull *github.Issue) error {
		if pull.Body != nil {
			cutie := GetCutieFromPull(pull)
			if cutie != nil {
				log.Println("Cutie:", *cutie)
				if err := tokens.twitter.PostToTwitter(cutie); err != nil {
					return err
				}
			}
		}
		return nil
	}

	lastPosted, err := tokens.twitter.LastPostedPull()
	if err != nil {
		log.Println(err)
		return
	}
	if lastPosted > 0 {
		tokens.github.PullsSinceFunc(lastPosted+1, TweetCutie)
	} else {
		tokens.github.PullsSinceFunc(StartCutiePullReq, TweetCutie)
	}
}

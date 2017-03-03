package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// TokensFile is file with credentials to access twitter and github
const TokensFile = "TOKENS"

// Tokens are the credentials for twitter and github
type Tokens struct {
	github  Github
	twitter Twitter
}

// LoadTokens loads tokens from file TokensFile
func LoadTokens() (*Tokens, error) {
	file, err := ioutil.ReadFile(TokensFile)
	if err != nil {
		return nil, err
	}
	tokens := &Tokens{}
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		token := strings.Fields(line)
		if len(token) > 0 {
			switch token[0] {
			case "twitterConsumerKey":
				tokens.twitter.twitterConsumerKey = token[2]
			case "twitterConsumerSecret":
				tokens.twitter.twitterConsumerSecret = token[2]
			case "twitterAccessToken":
				tokens.twitter.twitterAccessToken = token[2]
			case "twitterAccessSecret":
				tokens.twitter.twitterAccessSecret = token[2]
			case "githubPersonalAccessToken":
				tokens.github.githubPersonalAccessToken = token[2]
			default:
				return tokens, fmt.Errorf("Cannot identify token in line %q", line)
			}
		}
	}
	return tokens, nil
}

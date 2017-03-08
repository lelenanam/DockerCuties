package main

import (
	"regexp"

	"log"
	"strings"

	"github.com/google/go-github/github"
)

//StartCutiePullReq is the number of pull request
//where new template (with a picture of a cute animal) was added
const StartCutiePullReq = 20514

//Repo is repository with docker cuties
const Repo = "docker"

//Owner of repository
const Owner = "docker"

//SearchIssuesLimit is github limit for Search.Issues results
const SearchIssuesLimit = 1000

//PerPage number of pull requests per page
const PerPage = 50

//TwitterUser is account fot docker cuties in twitter
const TwitterUser = "DockerCuties"

//TwitterUploadLimit is limit for media upload in bytes
const TwitterUploadLimit = 3145728

// DockerCutie represents docker cutie by pull request URL and picture URL
type DockerCutie struct {
	pullnumber int
	pullURL    string
	cutieURL   string
}

// GetCutieFromPull parse body of pull request and return cutie if found link:
// ![image](https://cloud.githubusercontent.com/assets/2367858/23283487/02bb756e-f9db-11e6-9aa8-5f3e1bb80df3.png)
func GetCutieFromPull(pull *github.Issue) *DockerCutie {
	// TODO add flic.kr links, now just skip it
	if strings.Contains(*pull.Body, "flic.kr") {
		log.Println("flic.kr found")
		return nil
	}

	re := regexp.MustCompile(`!\[.*\]\((.*)\)`)
	result := re.FindStringSubmatch(*pull.Body)
	if len(result) > 1 {
		return &DockerCutie{
			pullnumber: *pull.Number,
			pullURL:    *pull.HTMLURL,
			cutieURL:   result[len(result)-1],
		}
	}
	return nil
}

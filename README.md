# DockerCuties

[![Build Status](https://travis-ci.org/lelenanam/DockerCuties.svg?branch=master)](https://travis-ci.org/lelenanam/DockerCuties)

Post pictures of cute animals from Docker's pull requests to twitter.

# Example

Collection of cuties from Docker pull requests on Twitter https://twitter.com/DockerCuties

# Usage

## Access tokens

To use this project you need credentials for accessing:
- Github https://github.com/blog/1509-personal-api-tokens
- Twitter https://dev.twitter.com/oauth/overview/application-owner-access-tokens

Add a file called `TOKENS` to your project. The format of this file should be:

```
twitterConsumerKey = <Twitter consumer key>
twitterConsumerSecret = <Twitter consumer secret>
twitterAccessToken = <Twitter access token>
twitterAccessSecret = <Twitter access token secret>
githubPersonalAccessToken = <Github personal access token>
```

## Docker image

Create image from Dockerfile with tag `cuteimage`:

```sh
$ docker build -t cuteimage github.com/lelenanam/DockerCuties
```

## Docker container

To check already running containers on the system use command:

```sh
$ docker ps
```

If there is previously ran container `cuteiner`, remove it:

```sh
$ docker rm -f cuteiner
```

Create and start a container in background (`-d`) from image `cuteimage`.
Bind mount volume `TOKENS`, assign the name `cuteiner` to the container.
To change log level use parameter `app --loglevel`

```sh
$ docker run -v $(pwd)/TOKENS:/go/src/app/TOKENS -d --name=cuteiner cuteimage --loglevel=info
```

Now you should see your `cuteiner` listed in the output for the `docker ps` command.

```sh
CONTAINER ID      IMAGE       COMMAND                  CREATED             STATUS           PORTS      NAMES
c111efc69db3      cuteimage   "app --loglevel=info"    14 seconds ago      Up 14 seconds               cuteiner
```

To fetch the logs of the container:

```sh
$ docker logs -f cuteiner
```

# Project Description

* Traverse all pull requests in Docker repository on Github and look for links in pull request body. The pull requests can be retrieved page by page.
* After the link is found try to get an image from URL.
* Check if the image is the screenshot. Don't post the screenshot and skip it.
* If the image is a picture, check the size of it because there is a twitter limit for media upload. Resize an image if needed.
* Encode image to string.
* Before posting new tweet we need to check the number of last posted pull request. Check most recent twitter timeline and find the largest number of posted pull request. 
* If there is no link in pull request just skip it. Otherwise, post it to twitter.
* Update twitter every 60 seconds for new cuties.
* There is a notification about screenshots and errors. There will be the direct message to twitter about it.

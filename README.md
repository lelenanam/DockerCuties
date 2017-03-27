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
Add a file called `TOKENS` to your project root. The format of this file should be:

```
twitterConsumerKey = <Twitter consumer key>
twitterConsumerSecret = <Twitter consumer secret>
twitterAccessToken = <Twitter access token>
twitterAccessSecret = <Twitter access token secret>
githubPersonalAccessToken = <Github personal access token>
```

## Docker image

Create image from Dockerfile with tag `cuties`:

```sh
$ docker build -t cuties github.com/lelenanam/DockerCuties
```

## Docker container

To check already running containers on the system use command:

```sh
$ docker ps
```

If there is previously ran container `cute`, remove it:

```sh
$ docker rm -f cute
```

Create and start a container in background (`-d`) from image `cuties`.
Bind mount volume `TOKENS`, assign the name `cute` to the container.
To change log level use parameter `app --loglevel`

```sh
$ docker run -v $(pwd)/TOKENS:/go/src/app/TOKENS -d --name=cute cuties app --loglevel=info
```

Now you should see your `cute` container listed in the output for the `docker ps` command.

```sh
CONTAINER ID        IMAGE          COMMAND                  CREATED             STATUS              PORTS         NAMES
c111efc69db3        cuties         "app --loglevel=info"    14 seconds ago      Up 14 seconds                     cute
```

To fetch the logs of the container:

```sh
$ docker logs -f cute
```

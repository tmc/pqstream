Contributing to pqstream
=====================

Welcome and thank you for taking time to contributing to the project.

## Menu

- [General](#general)
- [Making a pull request](#making_a_pull_request)
- [Setup your environment](#setup)

### General

- Check that your development setup is correct. (see #setup)

- Make sure your issue have not been attended to already by searching through the [Issues](https://github.com/tmc/pqstream/issues).

- Please follow Go best practices when making changes:
    - [Effective Go](https://golang.org/doc/effective_go.html)
    - [Code Review Comments](https://golang.org/wiki/CodeReviewComments)

- When comments are made about your changes, always assume positive intent.

### Making a pull request

Contributing to a Go project is slightly different because of import paths, please follow these steps to make it easier:

1. [Fork the repo](https://github.com/tmc/pqstream). This makes a copy of the code you can write to on your Github account. You will know have a repo called `pqstream` under your account i.e `https://github.com/<your_username>/pqstream`

2. If you haven't already done this, please `go get` the repo `go get github.com/tmc/pqstream`. This will download the repo into your local `$GOPATH`

3. Change directory to the downloaded repo and tell git that it can push to your fork by adding a remote: `git remote add fork https://github.com/<your_username>/pqstream.git`

4. Make your changes in the repo on your computer, preferably by branching

5. Push your changes to your fork: `git push fork`

6. Create a pull request to merge your changes into the `pqstream` **master** branch


**These guidelines are based on [GitHub and Go: forking, pull requests, and go-getting](http://blog.campoy.cat/2014/03/github-and-go-forking-pull-requests-and.html)**

### Setup

To be able to contribute to this project you will need make sure that you have the following installed on your development machine:

- Go environment  
    - [Install Go](https://golang.org/doc/install)
    - To setup your Go workspace, please read [How to write Go code](https://golang.org/doc/code.html)

- Protocol Buffers (TODO)
- gRPC (TODO)
- PostgreSQL (TODO)

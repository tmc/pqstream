Contributing to pqstream
========================

Welcome and thank you for taking time to contributing to the project.

## Menu

- [General](#general)
- [Making a pull request](#making-a-pull-request)
- [Reporting a bug](#report-a-bug)
- [Request a change/feature](#change-requests)
- [Setup your development environment](#setup)

### General

- Check that your development setup is correct. (see [setup](#setup))

- Make sure your issue have not been attended to already by searching through the project [Issues](https://github.com/tmc/pqstream/issues) page.

- Please follow Go best practices when making changes:
    - [Effective Go](https://golang.org/doc/effective_go.html)
    - [Code Review Comments](https://golang.org/wiki/CodeReviewComments)

- When comments are made about your changes, always assume positive intent.

### Making a pull request

Contributing to a Go project is slightly different because of import paths, please follow these steps to make it easier:

1. [Fork the repo](https://github.com/tmc/pqstream). This makes a copy of the code you can write to on your Github account. You will now have a repo called `pqstream` under your account i.e `https://github.com/<your_username>/pqstream`

2. If you haven't already done this, please `go get` the repo `go get github.com/tmc/pqstream`. This will download the repo into your local `$GOPATH`. The repo will be under `$GOPATH/src/github.com/tmc/pqstream`

3. Change directory to the downloaded repo (`cd $GOPATH/src/github.com/tmc/pqstream`) and tell git that it can push to your fork by adding a remote: `git remote add fork https://github.com/<your_username>/pqstream.git`

4. Make your changes in the repo on your computer, preferably by branching

5. Push your changes to your fork: `git push fork`

6. Create a pull request to merge your changes into the `pqstream` **master** branch


*These guidelines are based on [GitHub and Go: forking, pull requests, and go-getting](http://blog.campoy.cat/2014/03/github-and-go-forking-pull-requests-and.html)*

### Report a bug

Bugs can be reported by creating a new issue on the project [Issues](https://github.com/tmc/pqstream/issues) page.

### Change requests

Change requests can be logged by creating a new issue on the project [Issues](https://github.com/tmc/pqstream/issues) page. *Tip: Search through the issues first to see if someone else have to not maybe requested something similar.*

### Setup

To be able to contribute to this project you will need make sure that you have the following dependencies are installed on your development machine:

- Go environment (minimum version 1.7):
    - [Install Go](https://golang.org/doc/install)
    - To setup your Go workspace, please read [How to write Go code](https://golang.org/doc/code.html)

- Protocol Buffers 
    - The project uses protocol buffers. The `protoc` compiler can be downloaded from [here](https://github.com/google/protobuf/releases)

- The rest of the Go packages that are needed can be downloaded by running the `build` script in the root folder of the repo.



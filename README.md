# git-demo

A demo of [sshgate][sshgate] for Git push/pull

## Setup

- `go get github.com/xpensia/git-demo`
- `cd $GOPATH/src/github.com/xpensia/git-demo`
- Generate an ssh key with `ssh-keygen -t rsa` and save it in ./id_rsa

## Usage

`PORT=2222 go run app.go`


[sshgate]: https://github.com/xpensia/sshgate

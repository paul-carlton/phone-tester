
# Test Developers Guide

## Setup

clone repo:

    cd $(env gopath)/src/github.com/paul-carlton
    git clone git@github.com:paul-carlton/phone-tester.git
    cd phone-tester

## Development

The Makefile in the project's top level directory will compile, build and test all components.

    make docker

## Deployment

The Makefile can build a docker image too.

    make docker push

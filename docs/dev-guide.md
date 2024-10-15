
# Test Developers Guide

## Setup

clone repo:

    cd $(env gopath)/src/github.com/nabancard
    git clone git@github.com:nabancard/phone-tester.git
    cd phone-tester

Then build:

    make

## Development

The Makefile in the project's top level directory will compile, build and test all components.

    make build

## Deployment

The Makefile can build a docker image too.

    make push

# syntax=docker/dockerfile:experimental

FROM golang:1.22-bookworm as build

RUN apt-get update && apt-get install -y git

# See here for how we can clone git private repo: https://medium.com/@tonistiigi/build-secrets-and-ssh-forwarding-in-docker-18-09-ae8161d066
# Download public key for github.com
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN git config --global url."ssh://git@github.com/paul-carlton/".insteadOf "https://github.com/paul-carlton/" && \
    go env -w GOPRIVATE=github.com/paul-carlton,$(go env GOPRIVATE)

COPY . /opt/messaging

# Due to how docker workspaces work, I can't find a way to mount the host's
# /go/pkg/mod/cache dir into this container.
# So, allowing for a bit of a hack so the host can pull down all the modules once,
# and pass them into this. It requires the host to `cp -R  $GOPATH/pkg/mod/cache path-to/pos-api/tmp/go
# and then in here we extract that out to /go/pkg/mod/cache so that it does not try to download them
# which is important for CI.
RUN mkdir -p /opt/messaging/tmp/go/cache
RUN mkdir -p /go/pkg/mod/cache
RUN cp -R /opt/messaging/tmp/go/cache/ /go/pkg/mod/
RUN ls -lha /go/pkg/mod/cache
RUN --mount=type=ssh \
    ls /go/pkg/mod/cache && \
    cd /opt/messaging && make phone-tester-build && ls -l

FROM scratch

# Will need a ca-certificates.crt file for doing outgoing https requests
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# copy the app to root
COPY --from=build /go/bin/phone-tester .

# Command to execute
CMD [ "/phone-tester" ]

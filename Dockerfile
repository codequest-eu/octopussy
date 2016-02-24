FROM nanoservice/go:latest

# Create app directory.
ENV APP_HOME=/octopussy
RUN mkdir -p $APP_HOME

# Download dependencies.
RUN go get github.com/codegangsta/cli
RUN go get github.com/codequest-eu/octopussy/lib
RUN go get github.com/gorilla/websocket

# Add the code.
ADD . $APP_HOME

# Build dependencies and the octopussy binary.
WORKDIR $APP_HOME
RUN go build

# Clean up.
RUN apk del --purge go
RUN apk del --purge git
RUN rm -rf /go

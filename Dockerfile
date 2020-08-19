FROM golang:alpine AS builder 

#adding needed env variables 
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 
    
#move to /build 
WORKDIR /build 

#copy dependancies 
COPY go.mod . 
COPY go.sum . 
RUN go mod download 

#add code to container 
COPY . .

#build app 
RUN go build -o main . 

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/main .

# Build a small image
FROM scratch

COPY --from=builder /dist/main /

ENV DATABASE_URL=postgres://short:777777@lulu:5432/shorturl \ 
    GIN_MODE=release
    #change database url variable to match your needs 
    #comment out gin release mode to run in debug mode
    
# Command to run
ENTRYPOINT ["/main"]

EXPOSE 8080

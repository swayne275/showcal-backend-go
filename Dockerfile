FROM golang:alpine as build-stage

# TODO is git necessary once I switch to modules?
RUN apk --no-cache add ca-certificates git

WORKDIR /go/src/github.com/swayne275/showcal-backend-go

COPY . .

# Install dependencies
RUN go get "https://github.com/pkg/errors" \
	       "github.com/tidwall/gjson" \
	       "golang.org/x/net/context" \
	       "golang.org/x/oauth2" \
	       "golang.org/x/oauth2/google" \
	       "google.golang.org/api/calendar/v3" \
	       "google.golang.org/api/option"

RUN CGO_ENABLED=0 GOOS=linux go build -a -o /showcal-backend .

#
# final build stage
#
FROM scratch

# Copy ca-certs for application web access
COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-stage /showcal-backend /showcal-backend

# expose port 8080 to the outside world
EXPOSE 8080

ENTRYPOINT ["/showcal-backend"]
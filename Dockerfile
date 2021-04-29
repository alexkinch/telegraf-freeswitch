# From https://github.com/chemidy/smallest-secured-golang-docker-image/blob/master/docker/scratch_module.Dockerfile

ARG  BUILDER_IMAGE=golang:alpine
############################
# STEP 1 build executable binary
############################
FROM ${BUILDER_IMAGE} as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git openssh-client ca-certificates tzdata && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=1000

# RUN go env -w GOPRIVATE=github.com/telcoswitch

# See https://stackoverflow.com/a/55757473/12429735
RUN adduser \
	--disabled-password \
	--gecos "" \
	--home "/nonexistent" \
	--shell "/sbin/nologin" \
	--no-create-home \
	--uid "${UID}" \
	"${USER}"
WORKDIR $GOPATH/src/github.com/alexkinch/telegraf-freeswitch

# Use Go modules
COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
	-ldflags='-w -s -extldflags "-static"' -a \
	-o /go/bin/telegraf-freeswitch .

############################
# STEP 2 build a small image
############################
FROM scratch

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy our static executable
COPY --from=builder /go/bin/telegraf-freeswitch /go/bin/telegraf-freeswitch

# Use an unprivileged user.
USER appuser:appuser

# Run the binary.
ENTRYPOINT ["/go/bin/telegraf-freeswitch"]

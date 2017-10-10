FROM golang:1.9.1

ARG LDFLAGS
WORKDIR /go/src/github.com/linuxkit/rtf
COPY . /go/src/github.com/linuxkit/rtf

RUN CGO_ENABLED=0 go install -v -ldflags="$LDFLAGS" && \
	GOOS=darwin GOARCH=amd64 go install -v -ldflags="$LDFLAGS" && \
	GOOS=windows GOARCH=amd64 go install -v -ldflags="$LDFLAGS" && \
	GOOS=linux GOARCH=arm64 go install -v -ldflags="$LDFLAGS"
FROM scratch
COPY --from=0 /go/bin/rtf /rtf-Linux-x86_64
COPY --from=0 /go/bin/darwin_amd64/rtf /rtf-Darwin-x86_64
COPY --from=0 /go/bin/windows_amd64/rtf.exe /rtf-Windows-x86_64
COPY --from=0 /go/bin/linux_arm64/rtf /rtf-Linux-aarch64

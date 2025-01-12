FROM archlinux:latest
WORKDIR /app

RUN pacman -Syu --noconfirm &&  \
    pacman -S --noconfirm go meson ninja glib2 gtk3 gtk4 libadwaita wayland alsa-lib pulse-native-provider gobject-introspection graphene gcc pkgconf 

COPY cmd ./cmd
COPY internal ./internal
COPY misc ./misc
COPY go.* *.go default.pgo ./

ENV GO111MODULE=on
ENV CGO_ENABLED=1

RUN mkdir -p .bin
RUN go mod download
RUN go build -pgo default.pgo -tags wayland -trimpath -ldflags="-s -w" -o .bin/play-timer ./cmd/main.go 

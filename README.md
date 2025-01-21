# Play Timer 
[![Build](https://github.com/efogdev/mpris-timer/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/efogdev/mpris-timer/actions/workflows/build.yml)

Timer app pretending to be a media player. 

[![image](https://github.com/user-attachments/assets/75651dc5-de7a-4244-974a-47ee69adac0f)](https://flathub.org/apps/io.github.efogdev.mpris-timer)

Ultimately, serves the only purpose — to start a timer quickly and efficiently. \
Notifications included! Utilizing GTK4, Adwaita and MPRIS interface.

![image](https://github.com/user-attachments/assets/8f84bf5e-53a3-4919-a5b3-341b3f5f34b8)

## Installation

```shell
flatpak install flathub io.github.efogdev.mpris-timer
```

AUR: `play-timer` \
NixOS: [mpris-timer](https://github.com/NixOS/nixpkgs/blob/master/pkgs/by-name/mp/mpris-timer/package.nix) 

## Demo
![1](https://github.com/user-attachments/assets/9eab4435-9833-4f39-85e5-9a2f9ec3e75c)

Play Timer aims to be as keyboard friendly as possible. \
Use navigation keys (arrows, tab, shift+tab, space, enter) or start inputting numbers right away.

### KDE Plasma 
By default, Play Timer will use Breeze GTK4 theme with adjustments on Plasma. \
In case you prefer the original Adwaita look, set `PLAY_TIMER_IGNORE_KDE_THEME` environment variable to any non-empty value.

![image](https://github.com/user-attachments/assets/cc3f936e-c22f-4eb8-be7d-0c11e6b2228a)


## CLI use

```text
-ui
    Show timepicker UI (default true)
-start int
    Start the timer immediately, don't show UI (value in seconds)
-notify
    Send desktop notification (default true)
-rounded
    Rounded corners (default true)
-shadow
    Shadow for progress image
-color string
    Progress color (#HEX) for the player, use "default" for the GTK accent color (default "default")
-sound
    Play sound (default true)
-soundfile string
    Filename of the custom sound (must be .mp3)
-text string
    Notification text (default "Time is up!")
-title string
    Name/title of the timer (default "Timer")
-tray
    Force tray icon presence (default false)
-volume float
    Volume [0-1] (default 1)
-lowfps
    1 fps mode (energy saver, GNOME only)
```

#### Examples

```shell
# show UI for a red "Oven" timer
play-timer -title Oven -color "#FF4200"  

# start a silent 120s "Tea" timer immediately
play-timer -title Tea -rounded=0 -sound=0 -start 120
```

## Development

Install gsettings schema (the app will crash on start otherwise):
```shell
sudo cp misc/io.github.efogdev.mpris-timer.gschema.xml /usr/share/glib-2.0/schemas/
sudo glib-compile-schemas /usr/share/glib-2.0/schemas/
```

Run:
```shell
go run cmd/main.go -help
```

Build:
```shell
go build -pgo default.pgo -tags wayland -o ./.bin/app ./cmd/main.go
```
> There's a Dockerfile to build easily.

Flatpak:
```shell
flatpak run org.flatpak.Builder --force-clean --sandbox --user --install --install-deps-from=flathub --ccache .build io.github.efogdev.mpris-timer.yml
```

## Contributors

[@Tuba2](https://github.com/Tuba2) — app name and icon \
[@efogdev](https://github.com/efogdev) — oh wait that's me

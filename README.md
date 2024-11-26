# Play Timer
Timer app pretending to be a media player. 

[![image](https://github.com/user-attachments/assets/75651dc5-de7a-4244-974a-47ee69adac0f)](https://flathub.org/apps/io.github.efogdev.mpris-timer)

Ultimately, serves the only purpose — to start a timer quickly and efficiently. \
Notifications included! Utilizing GTK4, Adwaita and MPRIS interface.

![image](https://github.com/user-attachments/assets/8f84bf5e-53a3-4919-a5b3-341b3f5f34b8)

## Installation

```shell
flatpak install flathub io.github.efogdev.mpris-timer
```

## Demo
![2](https://github.com/user-attachments/assets/8fba3423-0133-4d79-8dfa-46c9995ba96b)

Play Timer aims to be as keyboard friendly as possible. \
Use navigation keys (arrows, tab, shift+tab, space, enter) or start inputting numbers right away.

## CLI use

```text
-ui
	Show timepicker UI (default true)
-start int
	Start the timer immediately, don't show UI (value in seconds)
-title string
	Name/title of the timer (default "Timer")
-text string
	Notification text (default "Time is up!")
-color string
	Progress color (#HEX) for the player, use "default" to use accent color (default "default")
-lowfps
	Low fps (~3 for KDE, ~15 for GNOME). On Plasma, FPS > 6 causes flickering in the media player widget. Some may experience this even with FPS <= 6 
-notify
	Send desktop notification (default true)
-rounded
	Rounded corners (default true)
-shadow
	Shadow for progress image
-silence int
	Play this milliseconds of silence before the actual sound — might be helpful for audio devices that wake up not immediately
-sound
	Play sound (default true)
-volume float
	Volume [0-1] (default 1)
```

### Examples

```shell
# show UI for a red "Oven" timer
flatpak run io.github.efogdev.mpris-timer -title Oven -color "#FF4200"  

# start a 120s "Tea" timer immediately
flatpak run io.github.efogdev.mpris-timer -title Tea -rounded=0 -sound=0 -start 120
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
go build -tags wayland -o ./.bin/app ./cmd/main.go
```

Flatpak:
```shell
flatpak run org.flatpak.Builder --force-clean --sandbox --user --install --install-deps-from=flathub --ccache .build io.github.efogdev.mpris-timer.yml
```

## ToDo

1) Custom sounds
2) Progress styles

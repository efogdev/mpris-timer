package ui

import (
	"context"
	"fmt"
	"github.com/diamondburned/gotk4/pkg/core/gioutil"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/efogdev/gotk4-adwaita/pkg/adw"
	"log"
	"mpris-timer/internal/core"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	prefsMinWidth      = 385
	prefsMinHeight     = 200
	prefsDefaultWidth  = 460
	prefsDefaultHeight = 660
	sliderWidth        = 175
)

var (
	prefsWin     *adw.Window
	previewImage *gtk.Image
)

func NewPrefsWindow() {
	if prefsWin != nil {
		return
	}

	prefsWin = adw.NewWindow()
	prefsWin.SetTitle("Preferences")
	prefsWin.SetSizeRequest(prefsMinWidth, prefsMinHeight)
	prefsWin.SetDefaultSize(prefsDefaultWidth, prefsDefaultHeight)

	prefsWin.ConnectCloseRequest(func() bool {
		prefsWin = nil
		return false
	})

	escCtrl := gtk.NewEventControllerKey()
	escCtrl.SetPropagationPhase(gtk.PhaseCapture)
	escCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		if slices.Contains(core.KeyEsc.GdkKeyvals(), keyval) {
			prefsWin.Close()
			return true
		}

		return false
	})

	header := adw.NewHeaderBar()
	view := adw.NewToolbarView()
	view.SetTopBarStyle(adw.ToolbarFlat)
	view.AddTopBar(header)

	box := gtk.NewBox(gtk.OrientationVertical, 8)
	box.SetVExpand(true)
	box.AddCSSClass("prefs-inner")

	NewPrefsWidgets(box)

	content := gtk.NewBox(gtk.OrientationVertical, 0)
	content.Append(box)

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetVExpand(true)
	scrolledWindow.SetHExpand(true)
	scrolledWindow.SetOverlayScrolling(true)
	scrolledWindow.SetChild(content)

	outerBox := gtk.NewBox(gtk.OrientationVertical, 0)
	outerBox.Append(header)
	outerBox.Append(scrolledWindow)

	view.SetContent(outerBox)
	prefsWin.AddController(escCtrl)
	prefsWin.SetContent(view)
	prefsWin.SetVisible(true)
	prefsWin.Activate()
	prefsWin.GrabFocus()

	// ToDo debug freezes w/o this
	<-time.After(20 * time.Millisecond)
	go renderPreview(previewImage)
}

var (
	presetsOnRightSwitch *adw.SwitchRow
	defaultPresetSelect  *adw.ComboRow
	presetsBox           *gtk.ListBox
)

func NewPrefsWidgets(parent *gtk.Box) {
	parent.SetSpacing(24)

	timerGroup := adw.NewPreferencesGroup()
	timerGroup.SetTitle("Timer")

	interfaceGroup := adw.NewPreferencesGroup()
	interfaceGroup.SetTitle("Interface")

	visualsGroup := adw.NewPreferencesGroup()
	visualsGroup.SetTitle("Visuals")
	visualsGroup.SetVAlign(gtk.AlignCenter)

	visualsBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	visualsBox.Append(visualsGroup)

	previewBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	previewBox.AddCSSClass("live-preview")
	previewBox.SetVAlign(gtk.AlignCenter)
	previewBox.SetHAlign(gtk.AlignEnd)
	previewBox.SetMarginTop(40)

	previewImage = gtk.NewImage()
	previewImage.SetSizeRequest(128, 128)

	previewBox.Append(previewImage)
	visualsBox.Append(previewBox)

	presetsGroup := adw.NewPreferencesGroup()
	presetsGroup.SetTitle("Interface")

	populateTimerGroup(timerGroup)
	populateInterfaceGroup(interfaceGroup)
	populateVisualsGroup(visualsGroup)
	populatePresetsGroup(presetsGroup)

	parent.Append(timerGroup)
	parent.Append(interfaceGroup)
	parent.Append(visualsBox)
	parent.Append(presetsGroup)
}

func populateInterfaceGroup(group *adw.PreferencesGroup) {
	titleEntry := adw.NewEntryRow()
	titleEntry.SetTitle("Default title")
	titleEntry.SetText(core.UserPrefs.DefaultTitle)
	titleEntry.ConnectChanged(func() {
		core.SetDefaultTitle(titleEntry.Text())
	})

	titleSwitch := adw.NewSwitchRow()
	titleSwitch.SetTitle("Title in UI")
	titleSwitch.SetSubtitle("Requires restart")
	titleSwitch.SetActive(core.UserPrefs.ShowTitle)
	titleSwitch.Connect("notify::active", func() {
		core.SetShowTitle(titleSwitch.Active())
	})

	forceTraySwitch := adw.NewSwitchRow()
	forceTraySwitch.SetTitle("Force tray icon")
	forceTraySwitch.SetActive(core.UserPrefs.ForceTrayIcon)
	forceTraySwitch.Connect("notify::active", func() {
		core.SetForceTrayIcon(forceTraySwitch.Active())
	})

	group.Add(titleEntry)
	group.Add(titleSwitch)
	group.Add(forceTraySwitch)
}

func populateTimerGroup(group *adw.PreferencesGroup) {
	textEntry := adw.NewEntryRow()
	volumeRow := adw.NewActionRow()
	volumeSlider := gtk.NewScaleWithRange(gtk.OrientationHorizontal, 0, 100, 1)
	customSoundSwitch := adw.NewSwitchRow()

	soundSwitch := adw.NewSwitchRow()
	soundSwitch.SetTitle("Enable sound")
	soundSwitch.SetSubtitle("Try clicking volume slider")
	soundSwitch.SetActive(core.UserPrefs.EnableSound)
	soundSwitch.Connect("notify::active", func() {
		core.SetEnableSound(soundSwitch.Active())
		volumeRow.SetVisible(core.UserPrefs.EnableSound)
		customSoundSwitch.SetVisible(core.UserPrefs.EnableSound)
	})

	customSoundSwitch = adw.NewSwitchRow()
	customSoundSwitch.SetTitle("Default sound")
	customSoundSwitch.SetVisible(core.UserPrefs.EnableSound)
	customSoundSwitch.SetActive(core.UserPrefs.SoundFilename == "")
	customSoundSwitch.Connect("notify::active", func() {
		if customSoundSwitch.Active() {
			core.SetSoundFilename("")
		} else {
			dialog := gtk.NewFileDialog()
			dialog.SetModal(true)
			dialog.SetTitle("MP3 sound")

			filter := gtk.NewFileFilter()
			filter.AddSuffix("mp3")
			filter.AddMIMEType("audio/mpeg")

			model := gioutil.NewListModel[*gtk.FileFilter]()
			model.Append(filter)

			dialog.SetFilters(model)
			dialog.SetDefaultFilter(filter)
			dialog.Open(context.Background(), &prefsWin.Window, func(r gio.AsyncResulter) {
				file, err := dialog.OpenFinish(r)
				if err == nil {
					core.SetSoundFilename(file.Path())
					_ = core.PlaySound()
				}
			})
		}
	})

	volumePreviewCtrl := gtk.NewGestureClick()
	volumePreviewCtrl.SetPropagationPhase(gtk.PhaseCapture)
	volumePreviewCtrl.ConnectReleased(func(_ int, _ float64, _ float64) {
		go func() { _ = core.PlaySound() }()
	})

	volumeRow.SetTitle("Sound volume")
	volumeRow.SetSubtitle(fmt.Sprintf("%v%%", int(core.Overrides.Volume*100)))
	volumeRow.SetVisible(core.UserPrefs.EnableSound)
	volumeRow.AddSuffix(volumeSlider)
	volumeRow.AddController(volumePreviewCtrl)

	volumeSlider.SetValue(core.Overrides.Volume * 100)
	volumeSlider.SetSizeRequest(sliderWidth, 0)
	volumeSlider.ConnectChangeValue(func(scroll gtk.ScrollType, value float64) (ok bool) {
		// GTK (probably) bug, scale goes up to 110 when using mouse wheel
		if value > 100 {
			value = 100
			volumeSlider.SetValue(value)
		}

		core.SetVolume(value / 100)
		volumeRow.SetSubtitle(fmt.Sprintf("%v%%", int(core.Overrides.Volume*100)))
		return false
	})

	notificationSwitch := adw.NewSwitchRow()
	notificationSwitch.SetTitle("Enable notification")
	notificationSwitch.SetActive(core.UserPrefs.ShouldNotify)
	notificationSwitch.Connect("notify::active", func() {
		core.SetEnableNotification(notificationSwitch.Active())
		textEntry.SetSensitive(notificationSwitch.Active())
	})

	textEntry.SetTitle("Default notification text")
	textEntry.SetText(core.UserPrefs.DefaultText)
	textEntry.SetSensitive(core.UserPrefs.ShouldNotify)
	textEntry.ConnectChanged(func() {
		core.SetDefaultText(textEntry.Text())
	})

	group.Add(soundSwitch)
	group.Add(customSoundSwitch)
	group.Add(volumeRow)
	group.Add(notificationSwitch)
	group.Add(textEntry)
}

func populateVisualsGroup(group *adw.PreferencesGroup) {
	color, err := core.RGBAFromHex(core.Overrides.Color)
	if err != nil {
		log.Fatalf("unexpected: nil color, %v (%s)", err, core.UserPrefs.ProgressColor)
	}

	dialog := gtk.NewColorDialog()
	dialog.SetWithAlpha(false)
	colorSwitch := gtk.NewColorDialogButton(dialog)
	colorSwitch.AddCSSClass("color-picker-btn")
	colorSwitch.SetRGBA(color)
	colorSwitch.SetVExpand(false)
	colorRow := adw.NewActionRow()
	colorRow.AddSuffix(colorSwitch)
	colorRow.SetTitle("Progress color")

	colorSwitch.Connect("notify", func() {
		core.SetProgressColor(core.HexFromRGBA(colorSwitch.RGBA()))
	})

	roundedSwitch := adw.NewSwitchRow()
	roundedSwitch.SetTitle("Rounded corners")
	roundedSwitch.SetActive(core.UserPrefs.Rounded)
	roundedSwitch.Connect("notify::active", func() {
		core.SetRounded(roundedSwitch.Active())
	})

	shadowSwitch := adw.NewSwitchRow()
	shadowSwitch.SetTitle("Shadow")
	shadowSwitch.SetActive(core.UserPrefs.Shadow)
	shadowSwitch.Connect("notify::active", func() {
		core.SetShadow(shadowSwitch.Active())
	})

	lowFPSSwitch := adw.NewSwitchRow()
	lowFPSSwitch.SetTitle("1 FPS mode")
	lowFPSSwitch.SetSubtitleLines(2)
	lowFPSSwitch.SetSubtitle(fmt.Sprintf("Does not affect preview"))
	lowFPSSwitch.SetHasTooltip(true)
	lowFPSSwitch.SetActive(core.UserPrefs.LowFPS)
	lowFPSSwitch.Connect("notify::active", func() {
		core.SetLowFPS(lowFPSSwitch.Active())
	})

	group.Add(colorRow)
	group.Add(roundedSwitch)

	if core.IsGnome || core.IsPlasma {
		group.Add(shadowSwitch)
	}

	if core.IsGnome {
		group.Add(lowFPSSwitch)
	}
}

func populatePresetsGroup(group *adw.PreferencesGroup) {
	defaultPresetSelect = adw.NewComboRow()
	newPresetBtn := gtk.NewButton()
	activatePresetSwitch := adw.NewSwitchRow()

	winSizeSwitch := adw.NewSwitchRow()
	winSizeSwitch.SetTitle("Remember window size")
	winSizeSwitch.SetActive(core.UserPrefs.RememberWinSize)
	winSizeSwitch.Connect("notify::active", func() {
		core.SetRememberWindowSize(winSizeSwitch.Active())
	})

	presetsOnRightSwitch = adw.NewSwitchRow()
	presetsOnRightSwitch.SetTitle("Presets on right side")
	presetsOnRightSwitch.SetSubtitle("Requires restart")
	presetsOnRightSwitch.SetSensitive(core.UserPrefs.ShowPresets)
	presetsOnRightSwitch.SetActive(core.UserPrefs.PresetsOnRight)
	presetsOnRightSwitch.Connect("notify::active", func() {
		core.SetPresetsOnRight(presetsOnRightSwitch.Active())
	})

	showPresetsSwitch := adw.NewSwitchRow()
	showPresetsSwitch.SetTitle("Show presets")
	showPresetsSwitch.SetSubtitle("Requires restart")
	showPresetsSwitch.SetActive(core.UserPrefs.ShowPresets)
	showPresetsSwitch.Connect("notify::active", func() {
		core.SetShowPresets(showPresetsSwitch.Active())
		presetsOnRightSwitch.SetSensitive(showPresetsSwitch.Active())
		defaultPresetSelect.SetSensitive(showPresetsSwitch.Active())
		activatePresetSwitch.SetSensitive(showPresetsSwitch.Active())
		presetsBox.SetVisible(showPresetsSwitch.Active())
		newPresetBtn.SetVisible(showPresetsSwitch.Active())
	})

	populateDefaultPresetSelect()
	defaultPresetSelect.SetTitle("Default preset")
	defaultPresetSelect.SetActivatable(true)
	defaultPresetSelect.SetSensitive(showPresetsSwitch.Active())
	defaultPresetSelect.Connect("notify::selected", func() {
		preset := core.UserPrefs.Presets[defaultPresetSelect.Selected()]
		core.SetDefaultPreset(preset)
	})

	activatePresetSwitch.SetTitle("Activate automatically")
	activatePresetSwitch.SetSensitive(core.UserPrefs.ShowPresets)
	activatePresetSwitch.SetActive(core.UserPrefs.ActivatePreset)
	activatePresetSwitch.Connect("notify::active", func() {
		core.SetActivatePreset(activatePresetSwitch.Active())
	})

	presetsBox = gtk.NewListBox()
	presetsBox.SetVisible(core.UserPrefs.ShowPresets)
	presetsBox.AddCSSClass("presets-list")
	presetsBox.SetVExpand(true)
	presetsBox.SetOverflow(gtk.OverflowHidden)

	RenderPresets([]string{})

	btnContent := adw.NewButtonContent()
	btnContent.SetLabel("")
	btnContent.SetIconName("list-add-symbolic")

	newPresetBtn.SetChild(btnContent)
	newPresetBtn.AddCSSClass("add-preset-btn")
	newPresetBtn.SetVisible(showPresetsSwitch.Active())
	newPresetBtn.ConnectClicked(func() {
		presets := append(core.UserPrefs.Presets, "00:00")
		core.SetPresets(presets)
		RenderPresets([]string{"00:00"})
	})

	footer := gtk.NewBox(gtk.OrientationHorizontal, 0)
	footer.SetHAlign(gtk.AlignCenter)
	footer.Append(newPresetBtn)

	group.Add(winSizeSwitch)
	group.Add(showPresetsSwitch)
	group.Add(presetsOnRightSwitch)
	group.Add(defaultPresetSelect)
	group.Add(activatePresetSwitch)
	group.Add(presetsBox)
	group.Add(footer)
}

func populateDefaultPresetSelect() {
	if len(core.UserPrefs.Presets) == 0 {
		defaultPresetSelect.SetModel(gtk.NewStringList(make([]string, 0)))
		return
	}

	selectedPos := slices.Index(core.UserPrefs.Presets, core.UserPrefs.DefaultPreset)
	i := uint(selectedPos)
	if selectedPos == -1 {
		i = 0
	}

	defaultPresetSelect.SetModel(gtk.NewStringList(core.UserPrefs.Presets))
	defaultPresetSelect.SetSelected(i)

	preset := core.UserPrefs.Presets[i]
	core.SetDefaultPreset(preset)
}

func RenderPresets(toAdd []string) {
	newPresets := toAdd
	if len(toAdd) == 0 {
		newPresets = core.UserPrefs.Presets
	}

	for _, preset := range newPresets {
		row := adw.NewActionRow()
		row.SetTitle(preset)

		container := gtk.NewListBoxRow()
		container.SetChild(row)

		box := gtk.NewBox(gtk.OrientationHorizontal, 16)
		box.SetVAlign(gtk.AlignCenter)
		box.AddCSSClass("presets-list-item")

		title := gtk.NewEditableLabel(preset)
		title.AddCSSClass("presets-list-title")
		title.SetVAlign(gtk.AlignStart)
		title.SetAlignment(0)
		title.SetHExpand(true)

		container.ConnectActivate(func() {
			title.GrabFocus()
			title.SelectRegion(0, -1)
		})

		cleanTitle := func() {
			title.GrabFocus()
			title.SetText("00:01")
		}

		formatTitle := func() {
			parts := strings.Split(title.Text(), ":")
			if len(parts) < 1 && len(parts) > 3 {
				cleanTitle()
				return
			}

			var vals []int
			for _, v := range parts {
				val, err := strconv.Atoi(v)
				if err != nil || val < 0 || val > 59 {
					cleanTitle()
					return
				}

				vals = append(vals, val)
			}

			var newText string
			if len(vals) == 1 {
				newText = fmt.Sprintf("00:%02d", vals[0])
			} else if len(vals) == 2 {
				newText = fmt.Sprintf("%02d:%02d", vals[0], vals[1])
			} else if len(vals) == 3 {
				newText = fmt.Sprintf("%02d:%02d:%02d", vals[0], vals[1], vals[2])
			}

			if newText == "" {
				cleanTitle()
				return
			}

			var presets []string
			for _idx, p := range core.UserPrefs.Presets {
				if container.Index() == _idx {
					presets = append(presets, newText)
					continue
				}

				presets = append(presets, p)
			}

			if container.Index() == slices.Index(core.UserPrefs.Presets, core.UserPrefs.DefaultPreset) {
				core.SetDefaultPreset(newText)
			}

			title.SetText(newText)
			core.SetPresets(presets)

			populateDefaultPresetSelect()
		}

		focusCtrl := gtk.NewEventControllerFocus()
		focusCtrl.ConnectLeave(formatTitle)
		title.AddController(focusCtrl)

		keyCtrl := gtk.NewEventControllerKey()
		keyCtrl.SetPropagationPhase(gtk.PhaseCapture)
		keyCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
			if slices.Contains(core.KeyEnter.GdkKeyvals(), keyval) {
				formatTitle()
			}

			return false
		})

		title.AddController(keyCtrl)

		btnContent := adw.NewButtonContent()
		btnContent.SetHExpand(false)
		btnContent.SetLabel("")
		btnContent.SetIconName("user-trash-symbolic")

		btn := gtk.NewButton()
		btn.SetChild(btnContent)
		btn.SetCursorFromName("pointer")
		btn.AddCSSClass("list-btn")
		btn.ConnectClicked(func() {
			var presets []string
			for _idx, p := range core.UserPrefs.Presets {
				if container.Index() == _idx {
					continue
				}

				presets = append(presets, p)
			}

			core.SetPresets(presets)
			presetsBox.Remove(container)
			populateDefaultPresetSelect()
		})

		box.Append(title)
		box.Append(btn)

		row.SetChild(box)
		row.SetActivatableWidget(title)
		presetsBox.Append(container)
	}
}

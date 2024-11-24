package ui

import (
	"fmt"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/efogdev/gotk4-adwaita/pkg/adw"
	"log"
	"mpris-timer/internal/util"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	prefsMinWidth      = 385
	prefsMinHeight     = 200
	prefsDefaultWidth  = 435
	prefsDefaultHeight = 625
	sliderWidth        = 175
)

var (
	prefsWin     *adw.Window
	previewImage *gtk.Image
)

func NewPrefsWindow() {
	prefsWin = adw.NewWindow()
	prefsWin.SetTitle("Preferences")
	prefsWin.SetSizeRequest(prefsMinWidth, prefsMinHeight)
	prefsWin.SetDefaultSize(prefsDefaultWidth, prefsDefaultHeight)

	escCtrl := gtk.NewEventControllerKey()
	escCtrl.SetPropagationPhase(gtk.PhaseCapture)
	escCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		if slices.Contains(util.KeyEsc.GdkKeyvals(), keyval) {
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
	populateVisualsGroup(visualsGroup)
	populatePresetsGroup(presetsGroup)

	parent.Append(timerGroup)
	parent.Append(visualsBox)
	parent.Append(presetsGroup)
}

func populateTimerGroup(group *adw.PreferencesGroup) {
	textEntry := adw.NewEntryRow()
	volumeRow := adw.NewActionRow()
	volumeSlider := gtk.NewScaleWithRange(gtk.OrientationHorizontal, 0, 100, 1)

	soundSwitch := adw.NewSwitchRow()
	soundSwitch.SetTitle("Enable sound")
	soundSwitch.SetActive(util.UserPrefs.EnableSound)
	soundSwitch.Connect("notify::active", func() {
		util.SetEnableSound(soundSwitch.Active())
		volumeRow.SetSensitive(util.UserPrefs.EnableSound)
	})

	volumePreviewCtrl := gtk.NewGestureClick()
	volumePreviewCtrl.SetPropagationPhase(gtk.PhaseCapture)
	volumePreviewCtrl.ConnectReleased(func(_ int, _ float64, _ float64) {
		go func() { _ = util.PlaySound(true) }()
	})

	volumeRow.SetTitle("Sound volume")
	volumeRow.SetSubtitle(fmt.Sprintf("%v%%", int(util.Overrides.Volume*100)))
	volumeRow.SetSensitive(util.UserPrefs.EnableSound)
	volumeRow.AddSuffix(volumeSlider)
	volumeRow.AddController(volumePreviewCtrl)

	volumeSlider.SetValue(util.Overrides.Volume * 100)
	volumeSlider.SetSizeRequest(sliderWidth, 0)
	volumeSlider.ConnectChangeValue(func(scroll gtk.ScrollType, value float64) (ok bool) {
		// GTK (probably) bug, scale goes up to 110 when using mouse wheel
		if value > 100 {
			value = 100
			volumeSlider.SetValue(value)
		}

		util.SetVolume(value / 100)
		volumeRow.SetSubtitle(fmt.Sprintf("%v%%", int(util.Overrides.Volume*100)))
		return false
	})

	notificationSwitch := adw.NewSwitchRow()
	notificationSwitch.SetTitle("Enable notification")
	notificationSwitch.SetActive(util.UserPrefs.EnableNotification)
	notificationSwitch.Connect("notify::active", func() {
		util.SetEnableNotification(notificationSwitch.Active())
		textEntry.SetSensitive(notificationSwitch.Active())
	})

	titleEntry := adw.NewEntryRow()
	titleEntry.SetTitle("Default title")
	titleEntry.SetText(util.UserPrefs.DefaultTitle)
	titleEntry.ConnectChanged(func() {
		util.SetDefaultTitle(titleEntry.Text())
	})

	titleSwitch := adw.NewSwitchRow()
	titleSwitch.SetTitle("Title in UI")
	titleSwitch.SetSubtitle("Requires restart")
	titleSwitch.SetActive(util.UserPrefs.ShowTitle)
	titleSwitch.Connect("notify::active", func() {
		util.SetShowTitle(titleSwitch.Active())
	})

	textEntry.SetTitle("Default notification text")
	textEntry.SetText(util.UserPrefs.DefaultText)
	textEntry.SetSensitive(util.UserPrefs.EnableNotification)
	textEntry.ConnectChanged(func() {
		util.SetDefaultText(textEntry.Text())
	})

	group.Add(soundSwitch)
	group.Add(volumeRow)
	group.Add(notificationSwitch)
	group.Add(titleEntry)
	group.Add(titleSwitch)
	group.Add(textEntry)
}

func populateVisualsGroup(group *adw.PreferencesGroup) {
	color, err := util.RGBAFromHex(util.Overrides.Color)
	if err != nil {
		log.Fatalf("unexpected: nil color, %v (%s)", err, util.UserPrefs.ProgressColor)
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
		util.SetProgressColor(util.HexFromRGBA(colorSwitch.RGBA()))
	})

	roundedSwitch := adw.NewSwitchRow()
	roundedSwitch.SetTitle("Rounded corners")
	roundedSwitch.SetActive(util.UserPrefs.Rounded)
	roundedSwitch.Connect("notify::active", func() {
		util.SetRounded(roundedSwitch.Active())
	})

	shadowSwitch := adw.NewSwitchRow()
	shadowSwitch.SetTitle("Shadow")
	shadowSwitch.SetActive(util.UserPrefs.Shadow)
	shadowSwitch.Connect("notify::active", func() {
		util.SetShadow(shadowSwitch.Active())
	})

	lowFPSSwitch := adw.NewSwitchRow()
	lowFPSSwitch.SetTitle("Lower FPS")
	lowFPSSwitch.SetSubtitleLines(2)
	lowFPSSwitch.SetSubtitle(fmt.Sprintf("Does not affect preview\nCurrently: %d FPS", util.CalculateFps()))
	lowFPSSwitch.SetHasTooltip(true)
	lowFPSSwitch.SetTooltipText("On Plasma, FPS > 6 causes flickering in the media player widget. Some may experience this even with FPS <= 6.")
	lowFPSSwitch.SetActive(util.UserPrefs.LowFPS)
	lowFPSSwitch.Connect("notify::active", func() {
		util.SetLowFPS(lowFPSSwitch.Active())
		lowFPSSwitch.SetSubtitle(fmt.Sprintf("Does not affect preview\nCurrently: %d FPS", util.CalculateFps()))
	})

	group.Add(colorRow)
	group.Add(roundedSwitch)
	group.Add(shadowSwitch)
	group.Add(lowFPSSwitch)
}

func populatePresetsGroup(group *adw.PreferencesGroup) {
	defaultPresetSelect = adw.NewComboRow()
	newPresetBtn := gtk.NewButton()
	activatePresetSwitch := adw.NewSwitchRow()

	winSizeSwitch := adw.NewSwitchRow()
	winSizeSwitch.SetTitle("Remember window size")
	winSizeSwitch.SetActive(util.UserPrefs.RememberWindowSize)
	winSizeSwitch.Connect("notify::active", func() {
		util.SetRememberWindowSize(winSizeSwitch.Active())
	})

	presetsOnRightSwitch = adw.NewSwitchRow()
	presetsOnRightSwitch.SetTitle("Presets on right side")
	presetsOnRightSwitch.SetSubtitle("Requires restart")
	presetsOnRightSwitch.SetSensitive(util.UserPrefs.ShowPresets)
	presetsOnRightSwitch.SetActive(util.UserPrefs.PresetsOnRight)
	presetsOnRightSwitch.Connect("notify::active", func() {
		util.SetPresetsOnRight(presetsOnRightSwitch.Active())
	})

	showPresetsSwitch := adw.NewSwitchRow()
	showPresetsSwitch.SetTitle("Show presets")
	showPresetsSwitch.SetSubtitle("Requires restart")
	showPresetsSwitch.SetActive(util.UserPrefs.ShowPresets)
	showPresetsSwitch.Connect("notify::active", func() {
		util.SetShowPresets(showPresetsSwitch.Active())
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
		preset := util.UserPrefs.Presets[defaultPresetSelect.Selected()]
		util.SetDefaultPreset(preset)
	})

	activatePresetSwitch.SetTitle("Activate automatically")
	activatePresetSwitch.SetSensitive(util.UserPrefs.ShowPresets)
	activatePresetSwitch.SetActive(util.UserPrefs.ActivatePreset)
	activatePresetSwitch.Connect("notify::active", func() {
		util.SetActivatePreset(activatePresetSwitch.Active())
	})

	presetsBox = gtk.NewListBox()
	presetsBox.SetVisible(util.UserPrefs.ShowPresets)
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
		presets := append(util.UserPrefs.Presets, "00:00")
		util.SetPresets(presets)
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
	if len(util.UserPrefs.Presets) == 0 {
		defaultPresetSelect.SetModel(gtk.NewStringList(make([]string, 0)))
		return
	}

	selectedPos := slices.Index(util.UserPrefs.Presets, util.UserPrefs.DefaultPreset)
	i := uint(selectedPos)
	if selectedPos == -1 {
		i = 0
	}

	defaultPresetSelect.SetModel(gtk.NewStringList(util.UserPrefs.Presets))
	defaultPresetSelect.SetSelected(i)

	preset := util.UserPrefs.Presets[i]
	util.SetDefaultPreset(preset)
}

func RenderPresets(toAdd []string) {
	newPresets := toAdd
	if len(toAdd) == 0 {
		newPresets = util.UserPrefs.Presets
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
			for _idx, p := range util.UserPrefs.Presets {
				if container.Index() == _idx {
					presets = append(presets, newText)
					continue
				}

				presets = append(presets, p)
			}

			if container.Index() == slices.Index(util.UserPrefs.Presets, util.UserPrefs.DefaultPreset) {
				util.SetDefaultPreset(newText)
			}

			title.SetText(newText)
			util.SetPresets(presets)

			populateDefaultPresetSelect()
		}

		focusCtrl := gtk.NewEventControllerFocus()
		focusCtrl.ConnectLeave(formatTitle)
		title.AddController(focusCtrl)

		keyCtrl := gtk.NewEventControllerKey()
		keyCtrl.SetPropagationPhase(gtk.PhaseCapture)
		keyCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
			if slices.Contains(util.KeyEnter.GdkKeyvals(), keyval) {
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
			for _idx, p := range util.UserPrefs.Presets {
				if container.Index() == _idx {
					continue
				}

				presets = append(presets, p)
			}

			util.SetPresets(presets)
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

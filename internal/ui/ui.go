package ui

import (
	_ "embed"
	"log"
	"mpris-timer/internal/core"
	"os"
	"slices"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/efogdev/gotk4-adwaita/pkg/adw"
)

//go:embed style/default.css
var cssString string

//go:embed style/breeze.css
var breezeCssString string

//go:embed res/icon.svg
var icon []byte

//go:embed res/icon.png
var iconPNG []byte

const (
	minWidth          = 350
	defaultMinHeight  = 195
	breezeMinHeight   = 165
	noTitleHeightDiff = 25
	collapseWidth     = 460
)

var (
	win           *adw.ApplicationWindow
	initialPreset *gtk.FlowBoxChild
	startBtn      *gtk.Button
	hrsLabel      *gtk.Entry
	minLabel      *gtk.Entry
	secLabel      *gtk.Entry
	titleLabel    *gtk.Entry
	flowBox       *gtk.FlowBox
	initComplete  bool
)

func Init() {
	gtk.Init()

	core.App.ConnectActivate(func() {
		prov := gtk.NewCSSProvider()
		prov.ConnectParsingError(func(sec *gtk.CSSSection, err error) {
			log.Printf("CSS error: %v", err)
		})

		allCss := cssString
		if core.BreezeTheme {
			allCss += breezeCssString
		}

		prov.LoadFromString(allCss)
		gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), prov, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
		NewTimePicker(core.App)
	})

	if code := core.App.Run(nil); code > 0 {
		os.Exit(code)
	}
}

func NewTimePicker(app *adw.Application) {
	core.Overrides.Duration = 0
	win = adw.NewApplicationWindow(&app.Application)
	handle := gtk.NewWindowHandle()
	body := adw.NewOverlaySplitView()
	handle.SetChild(body)

	escCtrl := gtk.NewEventControllerKey()
	escCtrl.SetPropagationPhase(gtk.PhaseCapture)
	escCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		isEsc := slices.Contains(core.KeyEsc.GdkKeyvals(), keyval)
		isCtrlQ := slices.Contains(core.KeyQ.GdkKeyvals(), keyval) && state == gdk.ControlMask
		isCtrlW := slices.Contains(core.KeyW.GdkKeyvals(), keyval) && state == gdk.ControlMask
		isCtrlD := slices.Contains(core.KeyD.GdkKeyvals(), keyval) && state == gdk.ControlMask

		if !isEsc && !isCtrlQ && !isCtrlW && !isCtrlD {
			return false
		}

		saveSize()
		win.Close()
		os.Exit(0)
		return true
	})

	win.AddController(escCtrl)
	win.SetContent(handle)
	win.SetTitle(core.AppName)
	win.SetSizeRequest(minWidth, getMinHeight())

	// ToDo refactor (at least magic numbers)
	width, height := int(core.UserPrefs.WindowWidth), int(core.UserPrefs.WindowHeight)
	if core.BreezeTheme && !core.UserPrefs.RememberWinSize {
		height -= 32
		width -= 60
	}

	win.SetDefaultSize(width, height)
	win.ConnectCloseRequest(func() (ok bool) {
		saveSize()
		return false
	})

	bp := adw.NewBreakpoint(adw.NewBreakpointConditionLength(adw.BreakpointConditionMaxWidth, collapseWidth, adw.LengthUnitSp))
	bp.AddSetter(body, "collapsed", true)
	win.AddBreakpoint(bp)

	body.SetVExpand(true)
	body.SetHExpand(true)

	if core.UserPrefs.PresetsOnRight {
		body.SetSidebarPosition(gtk.PackEnd)
	} else {
		body.SetSidebarPosition(gtk.PackStart)
	}

	body.SetContent(NewContent())
	body.SetSidebar(NewSidebar())
	body.SetSidebarWidthFraction(.36)
	body.SetEnableShowGesture(true)
	body.SetEnableHideGesture(true)
	body.SetShowSidebar(core.UserPrefs.ShowPresets && len(core.UserPrefs.Presets) > 0)
	body.SetMinSidebarWidth(40)

	win.AddCSSClass("root")
	win.SetVisible(true)
	minLabel.SetText("00")
	secLabel.SetText("00")

	if initialPreset != nil {
		initialPreset.Activate()
		initialPreset.GrabFocus()

		if !core.UserPrefs.ActivatePreset {
			minLabel.SetText("00")
			secLabel.SetText("00")
		} else {
			startBtn.GrabFocus()
		}
	}

	titleLabel.SetSensitive(true)
	win.Present()
	initComplete = true
}

func NewSidebar() *adw.NavigationPage {
	sidebar := adw.NewNavigationPage(gtk.NewBox(gtk.OrientationVertical, 0), "Presets")
	sidebar.SetOverflow(gtk.OverflowHidden)

	flowBox = gtk.NewFlowBox()
	flowBox.SetHomogeneous(true)
	flowBox.SetMinChildrenPerLine(1)
	flowBox.SetMaxChildrenPerLine(3)
	flowBox.SetSelectionMode(gtk.SelectionBrowse)
	flowBox.SetVAlign(gtk.AlignCenter)
	flowBox.SetColumnSpacing(16)
	flowBox.SetRowSpacing(16)
	flowBox.AddCSSClass("flow-box")

	for idx, preset := range core.UserPrefs.Presets {
		label := gtk.NewLabel(preset)
		label.SetCursorFromName("pointer")
		label.AddCSSClass("preset-lbl")
		label.SetHAlign(gtk.AlignCenter)
		label.SetVAlign(gtk.AlignCenter)
		flowBox.Append(label)

		onActivate := func() {
			time := core.TimeFromPreset(preset)

			if hrsLabel == nil || minLabel == nil || secLabel == nil {
				return
			}

			hrsLabel.SetText(core.NumToLabelText(time.Hour()))
			minLabel.SetText(core.NumToLabelText(time.Minute()))
			secLabel.SetText(core.NumToLabelText(time.Second()))

			if core.UserPrefs.StartPresetOnClick && initComplete {
				core.Overrides.Duration = time.Hour()*60*60 + time.Minute()*60 + time.Second()
				saveSize()
				win.Close()
				return
			} else {
				startBtn.SetCanFocus(true)
				startBtn.GrabFocus()
			}
		}

		mouseCtrl := gtk.NewGestureClick()
		mouseCtrl.ConnectReleased(func(nPress int, x, y float64) {
			onActivate()
		})

		child := flowBox.ChildAtIndex(idx)
		child.ConnectActivate(onActivate)
		child.AddController(mouseCtrl)

		if preset == core.UserPrefs.DefaultPreset {
			flowBox.SelectChild(child)
			initialPreset = child
		}

		keyCtrl := gtk.NewEventControllerKey()
		keyCtrl.SetPropagationPhase(gtk.PhaseCapture)
		keyCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
			if state != gdk.NoModifierMask {
				return false
			}

			// I don't like this solution but idk how to do it better
			x, _, w, _, _ := child.Bounds()

			if slices.Contains(core.KeyLeft.GdkKeyvals(), keyval) && x == 0 && core.UserPrefs.PresetsOnRight {
				secLabel.GrabFocus()
				return true
			}

			if slices.Contains(core.KeyRight.GdkKeyvals(), keyval) && (x+w == flowBox.Width()) && !core.UserPrefs.PresetsOnRight {
				minLabel.GrabFocus()
				return true
			}

			return false
		})

		child.AddController(keyCtrl)
	}

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetVExpand(true)
	scrolledWindow.SetOverlayScrolling(true)
	scrolledWindow.SetMinContentHeight(getMinHeight())
	scrolledWindow.SetChild(flowBox)

	kbCtrl := gtk.NewEventControllerKey()
	kbCtrl.SetPropagationPhase(gtk.PhaseBubble)
	kbCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		isNumber := core.IsGdkKeyvalNumber(keyval)
		if !isNumber {
			return false
		}

		minLabel.SetText(core.ParseKeyval(keyval))
		minLabel.Activate()
		minLabel.GrabFocus()
		minLabel.SelectRegion(1, 1)

		return true
	})

	sidebar.SetChild(scrolledWindow)
	sidebar.AddController(kbCtrl)

	return sidebar
}

func NewContent() *adw.NavigationPage {
	startBtn = gtk.NewButton()

	startFn := func() {
		time := core.TimeFromStrings(hrsLabel.Text(), minLabel.Text(), secLabel.Text())
		seconds := time.Hour()*60*60 + time.Minute()*60 + time.Second()
		if seconds > 0 {
			core.Overrides.Duration = seconds
			saveSize()
			win.Close()
		}
	}

	vBox := gtk.NewBox(gtk.OrientationVertical, 0)
	hBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	content := adw.NewNavigationPage(vBox, "New timer")
	vBox.SetHExpand(true)
	hBox.SetMarginStart(20)
	hBox.SetMarginEnd(20)

	titleLabel = gtk.NewEntry()
	titleLabel.SetHExpand(true)
	titleLabel.AddCSSClass("entry")
	titleLabel.AddCSSClass("title-entry")
	titleLabel.SetText(core.Overrides.Title)
	titleLabel.SetAlignment(.5)
	titleLabel.SetSensitive(false)

	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.SetPropagationPhase(gtk.PhaseCapture)
	keyCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		if state == gdk.NoModifierMask && slices.Contains(core.KeyEnter.GdkKeyvals(), keyval) {
			startFn()
			return true
		}

		_, pos, sel := titleLabel.SelectionBounds()
		if state == gdk.NoModifierMask && initialPreset != nil && !sel {
			toRight := core.UserPrefs.PresetsOnRight && slices.Contains(core.KeyRight.GdkKeyvals(), keyval) && pos == len(titleLabel.Text())
			toLeft := !core.UserPrefs.PresetsOnRight && slices.Contains(core.KeyLeft.GdkKeyvals(), keyval) && pos == 0
			if toRight || toLeft {
				initialPreset.GrabFocus()
				return true
			}
		}

		return false
	})

	titleLabel.AddController(keyCtrl)
	titleLabel.ConnectChanged(func() {
		core.Overrides.Title = titleLabel.Text()
	})

	titleBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	titleBox.AddCSSClass("title-box")
	titleBox.SetVAlign(gtk.AlignCenter)
	titleBox.SetHExpand(true)
	titleBox.Append(titleLabel)

	if core.UserPrefs.ShowTitle {
		vBox.Append(titleBox)
	}
	vBox.Append(hBox)

	hrsLabel = gtk.NewEntry()
	minLabel = gtk.NewEntry()
	secLabel = gtk.NewEntry()

	fin := func() { startBtn.Activate() }
	setupTimeEntry(hrsLabel, nil, &minLabel.Widget, 23, fin)
	setupTimeEntry(minLabel, &hrsLabel.Widget, &secLabel.Widget, 59, fin)
	setupTimeEntry(secLabel, &minLabel.Widget, &startBtn.Widget, 59, fin)

	hrsLeftCtrl := gtk.NewEventControllerKey()
	hrsLeftCtrl.SetPropagationPhase(gtk.PhaseCapture)
	hrsLeftCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		selected := flowBox.SelectedChildren()

		if len(selected) != 1 {
			return false
		}

		if !core.UserPrefs.PresetsOnRight && slices.Contains(core.KeyLeft.GdkKeyvals(), keyval) {
			selected[0].GrabFocus()
		}

		return false
	})

	hrsLabel.AddController(hrsLeftCtrl)

	scLabel1 := gtk.NewLabel(":")
	scLabel1.AddCSSClass("semicolon")

	scLabel2 := gtk.NewLabel(":")
	scLabel2.AddCSSClass("semicolon")

	hBox.Append(hrsLabel)
	hBox.Append(scLabel1)
	hBox.Append(minLabel)
	hBox.Append(scLabel2)
	hBox.Append(secLabel)

	hBox.SetVAlign(gtk.AlignCenter)
	hBox.SetHAlign(gtk.AlignCenter)
	hBox.SetVExpand(true)
	hBox.SetHExpand(true)

	btnContent := adw.NewButtonContent()
	btnContent.SetHExpand(false)
	btnContent.SetLabel("Start")
	btnContent.SetIconName("media-playback-start-symbolic")

	startBtn.SetCanFocus(false)
	startBtn.SetChild(btnContent)
	startBtn.SetHExpand(false)
	startBtn.AddCSSClass("control-btn")
	startBtn.AddCSSClass("suggested-action")

	leftKeyCtrl := gtk.NewEventControllerKey()
	leftKeyCtrl.SetPropagationPhase(gtk.PhaseCapture)
	leftKeyCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		if slices.Contains(core.KeyLeft.GdkKeyvals(), keyval) && state == gdk.NoModifierMask {
			secLabel.GrabFocus()
			return true
		}

		return false
	})

	startBtn.ConnectClicked(startFn)
	startBtn.ConnectActivate(startFn)
	startBtn.AddController(leftKeyCtrl)

	prefsBtnContent := adw.NewButtonContent()
	prefsBtnContent.SetHExpand(false)
	prefsBtnContent.SetLabel("")
	prefsBtnContent.SetIconName("emblem-system-symbolic")

	prefsBtn := gtk.NewButton()
	prefsBtn.SetTooltipText("Preferences")
	prefsBtn.SetChild(prefsBtnContent)
	prefsBtn.AddCSSClass("control-btn")
	prefsBtn.AddCSSClass("prefs-btn")
	prefsBtn.SetFocusable(false)
	prefsBtn.ConnectClicked(func() {
		NewPrefsWindow()
	})

	closeBtnContent := adw.NewButtonContent()
	closeBtnContent.SetHExpand(false)
	closeBtnContent.SetLabel("")
	closeBtnContent.SetIconName("application-exit-symbolic")

	exitBtn := gtk.NewButton()
	exitBtn.SetTooltipText("Exit")
	exitBtn.SetChild(closeBtnContent)
	exitBtn.AddCSSClass("control-btn")
	exitBtn.AddCSSClass("prefs-btn")
	exitBtn.SetFocusable(false)
	exitBtn.ConnectClicked(func() {
		win.Close()
		os.Exit(0)
	})

	footer := gtk.NewBox(gtk.OrientationHorizontal, 12)
	footer.SetVAlign(gtk.AlignCenter)
	footer.SetHAlign(gtk.AlignCenter)
	footer.SetHExpand(false)
	footer.SetMarginBottom(4)
	footer.AddCSSClass("footer")
	footer.Append(startBtn)
	footer.Append(prefsBtn)
	footer.Append(exitBtn)
	vBox.Append(footer)

	return content
}

func getMinHeight() int {
	height := defaultMinHeight
	if !core.UserPrefs.ShowTitle {
		height -= noTitleHeightDiff
	}
	if core.BreezeTheme {
		height = breezeMinHeight
	}

	return height
}

func saveSize() {
	if core.UserPrefs.RememberWinSize {
		core.SetWindowSize(uint(win.Width()), uint(win.Height()))
	}
}

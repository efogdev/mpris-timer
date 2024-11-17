package ui

import (
	_ "embed"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/efogdev/gotk4-adwaita/pkg/adw"
	"log"
	"mpris-timer/internal/util"
	"os"
	"slices"
)

//go:embed style.css
var cssString string

const (
	minWidth      = 400
	minHeight     = 160
	collapseWidth = 520
)

var (
	win           *adw.ApplicationWindow
	initialPreset *gtk.FlowBoxChild
	startBtn      *gtk.Button
	hrsLabel      *gtk.Entry
	minLabel      *gtk.Entry
	secLabel      *gtk.Entry
	flowBox       *gtk.FlowBox
)

func Init() {
	log.Println("UI init requested")

	util.App.ConnectActivate(func() {
		prov := gtk.NewCSSProvider()
		prov.ConnectParsingError(func(sec *gtk.CSSSection, err error) {
			log.Printf("CSS error: %v", err)
		})

		prov.LoadFromString(cssString)
		gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), prov, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

		NewTimePicker(util.App)
	})

	if code := util.App.Run(nil); code > 0 {
		os.Exit(code)
	}
}

func NewTimePicker(app *adw.Application) {
	util.Overrides.Duration = 0
	win = adw.NewApplicationWindow(&app.Application)
	handle := gtk.NewWindowHandle()
	body := adw.NewOverlaySplitView()
	handle.SetChild(body)

	escCtrl := gtk.NewEventControllerKey()
	escCtrl.SetPropagationPhase(gtk.PhaseCapture)
	escCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		if !slices.Contains(util.KeyEsc.GdkKeyvals(), keyval) {
			return false
		}

		saveSize()
		win.Close()
		os.Exit(0)
		return true
	})

	win.AddController(escCtrl)
	win.SetContent(handle)
	win.SetTitle(util.AppName)
	win.SetSizeRequest(minWidth, minHeight)
	win.SetDefaultSize(int(util.UserPrefs.WindowWidth), int(util.UserPrefs.WindowHeight))

	win.ConnectCloseRequest(func() (ok bool) {
		saveSize()
		return false
	})

	bp := adw.NewBreakpoint(adw.NewBreakpointConditionLength(adw.BreakpointConditionMaxWidth, collapseWidth, adw.LengthUnitSp))
	bp.AddSetter(body, "collapsed", true)
	win.AddBreakpoint(bp)

	body.SetVExpand(true)
	body.SetHExpand(true)

	if util.UserPrefs.PresetsOnRight {
		body.SetSidebarPosition(gtk.PackEnd)
	} else {
		body.SetSidebarPosition(gtk.PackStart)
	}

	body.SetContent(NewContent())
	body.SetSidebar(NewSidebar())
	body.SetSidebarWidthFraction(.35)
	body.SetEnableShowGesture(true)
	body.SetEnableHideGesture(true)
	body.SetShowSidebar(util.UserPrefs.ShowPresets && len(util.UserPrefs.Presets) > 0)

	win.SetVisible(true)
	minLabel.SetText("00")
	secLabel.SetText("00")

	if initialPreset != nil {
		initialPreset.Activate()
		initialPreset.GrabFocus()

		if !util.UserPrefs.ActivatePreset {
			minLabel.SetText("00")
			secLabel.SetText("00")
		} else {
			startBtn.GrabFocus()
		}
	}

	win.Present()
}

func NewSidebar() *adw.NavigationPage {
	sidebar := adw.NewNavigationPage(gtk.NewBox(gtk.OrientationVertical, 0), "Presets")
	flowBox = gtk.NewFlowBox()

	flowBox.SetSelectionMode(gtk.SelectionBrowse)
	flowBox.SetVAlign(gtk.AlignCenter)
	flowBox.SetColumnSpacing(16)
	flowBox.SetRowSpacing(16)
	flowBox.AddCSSClass("flow-box")

	for idx, preset := range util.UserPrefs.Presets {
		label := gtk.NewLabel(preset)
		label.SetCursorFromName("pointer")
		label.AddCSSClass("preset-lbl")
		label.SetHAlign(gtk.AlignCenter)
		label.SetVAlign(gtk.AlignCenter)
		flowBox.Append(label)

		onActivate := func() {
			time := util.TimeFromPreset(preset)

			if hrsLabel == nil || minLabel == nil || secLabel == nil {
				return
			}

			hrsLabel.SetText(util.NumToLabelText(time.Hour()))
			minLabel.SetText(util.NumToLabelText(time.Minute()))
			secLabel.SetText(util.NumToLabelText(time.Second()))
			startBtn.SetCanFocus(true)
			startBtn.GrabFocus()
		}

		mouseCtrl := gtk.NewGestureClick()
		mouseCtrl.ConnectReleased(func(nPress int, x, y float64) {
			onActivate()
		})

		child := flowBox.ChildAtIndex(idx)
		child.ConnectActivate(onActivate)
		child.AddController(mouseCtrl)

		if preset == util.UserPrefs.DefaultPreset {
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

			if slices.Contains(util.KeyLeft.GdkKeyvals(), keyval) && x == 0 && util.UserPrefs.PresetsOnRight {
				secLabel.GrabFocus()
				return true
			}

			if slices.Contains(util.KeyRight.GdkKeyvals(), keyval) && (x+w == flowBox.Width()) && !util.UserPrefs.PresetsOnRight {
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
	scrolledWindow.SetMinContentHeight(minHeight)
	scrolledWindow.SetChild(flowBox)

	kbCtrl := gtk.NewEventControllerKey()
	kbCtrl.SetPropagationPhase(gtk.PhaseBubble)
	kbCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		isNumber := util.IsGdkKeyvalNumber(keyval)
		if !isNumber {
			return false
		}

		minLabel.SetText(util.ParseKeyval(keyval))
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

	vBox := gtk.NewBox(gtk.OrientationVertical, 8)
	hBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	clamp := adw.NewClamp()
	content := adw.NewNavigationPage(clamp, "New timer")

	clamp.SetChild(vBox)
	vBox.Append(hBox)

	hrsLabel = gtk.NewEntry()
	minLabel = gtk.NewEntry()
	secLabel = gtk.NewEntry()

	finish := func() {
		startBtn.Activate()
	}

	setupTimeEntry(hrsLabel, nil, &minLabel.Widget, 23, finish)
	setupTimeEntry(minLabel, &hrsLabel.Widget, &secLabel.Widget, 59, finish)
	setupTimeEntry(secLabel, &minLabel.Widget, &startBtn.Widget, 59, finish)

	hrsLeftCtrl := gtk.NewEventControllerKey()
	hrsLeftCtrl.SetPropagationPhase(gtk.PhaseCapture)
	hrsLeftCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		selected := flowBox.SelectedChildren()

		if len(selected) != 1 {
			return false
		}

		if !util.UserPrefs.PresetsOnRight && slices.Contains(util.KeyLeft.GdkKeyvals(), keyval) {
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

	startFn := func() {
		time := util.TimeFromStrings(hrsLabel.Text(), minLabel.Text(), secLabel.Text())
		seconds := time.Hour()*60*60 + time.Minute()*60 + time.Second()
		if seconds > 0 {
			util.Overrides.Duration = seconds
			saveSize()
			win.Close()
			return
		}

		os.Exit(1)
	}

	leftKeyCtrl := gtk.NewEventControllerKey()
	leftKeyCtrl.SetPropagationPhase(gtk.PhaseCapture)
	leftKeyCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) (ok bool) {
		if slices.Contains(util.KeyLeft.GdkKeyvals(), keyval) && state == gdk.NoModifierMask {
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
	prefsBtnContent.SetIconName("preferences-system-symbolic")

	prefsBtn := gtk.NewButton()
	prefsBtn.SetChild(prefsBtnContent)
	prefsBtn.AddCSSClass("control-btn")
	prefsBtn.AddCSSClass("prefs-btn")
	prefsBtn.SetFocusable(false)

	prefsBtn.ConnectClicked(func() {
		NewPrefsWindow()
	})

	footer := gtk.NewBox(gtk.OrientationHorizontal, 16)
	footer.SetVAlign(gtk.AlignCenter)
	footer.SetHAlign(gtk.AlignCenter)
	footer.SetHExpand(false)
	footer.SetMarginBottom(16)
	footer.AddCSSClass("footer")
	footer.Append(startBtn)
	footer.Append(prefsBtn)
	vBox.Append(footer)

	return content
}

func saveSize() {
	if util.UserPrefs.RememberWindowSize {
		util.SetWindowSize(uint(win.Width()), uint(win.Height()))
	}
}

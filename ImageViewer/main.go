package main

import (
	"fmt"
	"net/url"
	"os"
    "path/filepath"  // Is this for generating windows-friendly filepaths?

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
    "github.com/spf13/viper"
)

const (
    // tagBtnTotal must match size of DefaultButtonTags()
    tagBtnTotal = 30
    viperFilename = "config"
)

func viperPath() string {
    return filepath.Join(os.Getenv("HOME"), ".imagetagger")
}

func DefaultButtonTags() []string {
    return []string {
        "AC",
        "AC DATA",
        "ATTIC",
        "BASEMENT",
        "BOILER",
        "BOILER DATA",
        "BLOWER",
        "BUFFER",
        "FRONT",
        "FURNACE",
        "FURNACE DATA",
        "GARAGE",
        "GAS",
        "GRADE",
        "HEADER",
        "HP",
        "HP DATA",
        "HRV",
        "HRV DATA",
        "INSUL",
        "LEFT",
        "PONY",
        "REAR",
        "RIGHT",
        "TANK",
        "TANK DATA",
        "TAX",
        "",
        "",
        "",
    }
}

type enterEntry struct {
	widget.Entry
	enterFunc func(s string)
}

func (e *enterEntry) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyReturn:
		e.enterFunc(e.Text)
	default:
		e.Entry.TypedKey(key)
	}
}

func newEnterEntry() *enterEntry {
	entry := &enterEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

type editingSlider struct {
	widget.Slider
	dragEndFunc func(float64)
}

func (e *editingSlider) DragEnd() {
	e.dragEndFunc(e.Value)
}

func newEditingSlider(min, max float64) *editingSlider {
	editSlider := &editingSlider{}
	editSlider.Max = max
	editSlider.Min = min
	editSlider.ExtendBaseWidget(editSlider)
	return editSlider
}

// newEditingOption creates a new VBox, that includes an info text and a widget to edit the parameter
func newEditingOption(infoText string, slider *editingSlider, defaultValue float64) *fyne.Container {
	data := binding.BindFloat(&defaultValue)
	text := widget.NewLabel(infoText)
	value := widget.NewLabelWithData(binding.FloatToStringWithFormat(data, "%.0f"))
	slider.Bind(data)
	slider.Step = 1

	return container.NewVBox(
		container.NewHBox(
			text,
			layout.NewSpacer(),
			value,
		),
		slider,
	)
}

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}
	return link
}

// App represents the whole application with all its windows, widgets and functions
type App struct {
	app     fyne.App
	mainWin fyne.Window
    config  *viper.Viper

	img        Img
	mainModKey desktop.Modifier
	focus      bool
	lastOpened []string
    file       *os.File

	image *canvas.Image

	sliderBrightness    *editingSlider
	sliderContrast      *editingSlider
	sliderHue           *editingSlider
	sliderSaturation    *editingSlider
	sliderColorBalanceR *editingSlider
	sliderColorBalanceG *editingSlider
	sliderColorBalanceB *editingSlider
	sliderSepia         *editingSlider
	sliderBlur          *editingSlider
	resetBtn            *widget.Button

	split       *container.Split
	widthLabel  *widget.Label
	heightLabel *widget.Label
	imgSize     *widget.Label
	imgLastMod  *widget.Label
	tagBtnLabel *widget.Label
    tagBtns     []*widget.Button
    tagBtnEntries   []*widget.Entry
    tagBtnGrid  *fyne.Container
    saveTagsBtn *widget.Button
    editTagsBtn *widget.Button

	bottomBar       *fyne.Container
	bottomBarSplit  *container.Split
    renamePreview   *widget.Entry
	leftArrow       *widget.Button
	rightArrow      *widget.Button
	confirmArrow    *widget.Button

	statusBar    *fyne.Container
	deleteBtn    *widget.Button
	renameBtn    *widget.Button
	zoomIn       *widget.Button
	zoomOut      *widget.Button
	zoomLabel    *widget.Label
	resetZoomBtn *widget.Button

	fullscreenWin fyne.Window
}

func reverseArray(arr []string) []string {
	for i := 0; i < len(arr)/2; i++ {
		j := len(arr) - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

func (a *App) init() {
	a.img = Img{}
	a.img.init()

	// theme
	switch a.app.Preferences().StringWithFallback("Theme", "Dark") {
	case "Light":
		a.app.Settings().SetTheme(theme.LightTheme())
	case "Dark":
		a.app.Settings().SetTheme(theme.DarkTheme())
	}

	// show/hide statusbar
	if a.app.Preferences().BoolWithFallback("statusBarVisible", true) == false {
		a.statusBar.Hide()
	}
}

func (a *App) WriteConfig() {
    if err := os.MkdirAll(viperPath(), os.ModePerm); err != nil {
        fmt.Errorf("Error creating config file directory: %w. Default configs will be used.\n", err)
    }
    if err := a.config.WriteConfigAs(filepath.Join(viperPath(), viperFilename)); err != nil {
        fmt.Errorf("Error writing config file: %w. Updates to configs will not be saved.\n", err)
    }
}

func main() {
    viperConfig := viper.New()

    viperConfig.SetDefault("auto-generated-file", "This file managed by Image Tagger. Do not modify!")
    viperConfig.SetDefault("ImagePath", os.Getenv("HOME"))
    viperConfig.SetDefault("ButtonTags", DefaultButtonTags() )

    viperConfig.SetConfigName(viperFilename)       // name of config file (without extension)
    viperConfig.SetConfigType("yaml")
    viperConfig.AddConfigPath(viperPath())

    if err := viperConfig.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // Config file not found; use defaults.
            fmt.Printf("No config file found; defaults will be used.\n")
        } else {
            // Config file was found but another error was produced
            fmt.Errorf("Error reading config: %w. Default configs will be used.\n", err)
        }
    }

	a := app.NewWithID("io.github.jjwinters.image-tagger")
	w := a.NewWindow("Image Tagger")
	a.SetIcon(resourceIconPng)
	w.SetIcon(resourceIconPng)
	ui := &App{app: a, mainWin: w, config: viperConfig}
	ui.init()
    ui.WriteConfig()
	w.SetContent(ui.loadMainUI())
	if len(os.Args) > 1 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Printf("error while opening the file: %v\n", err)
		}
		ui.open(file, true)
	}
	w.Resize(fyne.NewSize(1200, 750))
	w.ShowAndRun()
}


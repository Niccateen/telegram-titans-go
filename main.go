package main

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Custom purple theme
type purpleTheme struct{}

func (p purpleTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 45, G: 25, B: 70, A: 255} // Deep purple background
	case theme.ColorNameForeground:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255} // Pure white text
	case theme.ColorNameButton:
		return color.RGBA{R: 90, G: 50, B: 140, A: 255} // Purple buttons
	case theme.ColorNamePrimary:
		return color.RGBA{R: 180, G: 130, B: 255, A: 255} // Light purple accent
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 60, G: 35, B: 95, A: 255} // Slightly lighter purple for inputs
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 180, G: 170, B: 200, A: 255} // Light purple placeholder
	case theme.ColorNameScrollBar:
		return color.RGBA{R: 120, G: 80, B: 180, A: 255}
	case theme.ColorNameSeparator:
		return color.RGBA{R: 100, G: 70, B: 150, A: 255}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 150, G: 140, B: 170, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (p purpleTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (p purpleTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (p purpleTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// ANSI escape code regex - strips colors, cursor movement, QR codes, etc.
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\].*?\x07|\x1b[()][AB012]|\x1b\[[\?0-9;]*[a-zA-Z]`)

// App holds the application state
type App struct {
	fyneApp    fyne.App
	mainWindow fyne.Window
	outputArea *widget.Entry
	outputMu   sync.Mutex

	// Settings
	namespace string
	proxy     string

	// Current view
	contentContainer *fyne.Container
}

// stripANSI removes all ANSI escape sequences from text
func stripANSI(text string) string {
	return ansiRegex.ReplaceAllString(text, "")
}

func main() {
	a := &App{
		fyneApp:   app.NewWithID("com.tdl.gui"),
		namespace: "default",
		proxy:     "",
	}

	// Use custom purple theme with white text
	a.fyneApp.Settings().SetTheme(&purpleTheme{})

	a.mainWindow = a.fyneApp.NewWindow("TDL Manager")
	a.mainWindow.Resize(fyne.NewSize(700, 600))

	// Create output area (shared across all screens)
	a.outputArea = widget.NewMultiLineEntry()
	a.outputArea.SetPlaceHolder("Command output will appear here...")
	a.outputArea.Wrapping = fyne.TextWrapWord
	a.outputArea.Disable()

	// Create the main UI
	a.showMainScreen()

	a.mainWindow.ShowAndRun()
}

// appendOutput safely appends text to the output area
func (a *App) appendOutput(text string) {
	a.outputMu.Lock()
	defer a.outputMu.Unlock()
	current := a.outputArea.Text
	if current == "" {
		a.outputArea.SetText(text)
	} else {
		a.outputArea.SetText(current + "\n" + text)
	}
	// Scroll to bottom
	a.outputArea.CursorRow = len(strings.Split(a.outputArea.Text, "\n")) - 1
}

// clearOutput clears the output area
func (a *App) clearOutput() {
	a.outputMu.Lock()
	defer a.outputMu.Unlock()
	a.outputArea.SetText("")
}

// runCommand executes a tdl command and streams output to the output area
func (a *App) runCommand(args ...string) {
	a.clearOutput()
	a.appendOutput(fmt.Sprintf("$ tdl %s", strings.Join(args, " ")))
	a.appendOutput("")

	// Prepend global flags
	finalArgs := []string{}
	if a.namespace != "" && a.namespace != "default" {
		finalArgs = append(finalArgs, "-n", a.namespace)
	}
	if a.proxy != "" {
		finalArgs = append(finalArgs, "--proxy", a.proxy)
	}
	finalArgs = append(finalArgs, args...)

	cmd := exec.Command("tdl", finalArgs...)

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.appendOutput(fmt.Sprintf("Error creating stdout pipe: %v", err))
		return
	}

	// Get stderr pipe
	stderr, err := cmd.StderrPipe()
	if err != nil {
		a.appendOutput(fmt.Sprintf("Error creating stderr pipe: %v", err))
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		a.appendOutput(fmt.Sprintf("Error starting command: %v", err))
		return
	}

	// Read stdout in a goroutine
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		a.streamOutput(stdout)
	}()

	go func() {
		defer wg.Done()
		a.streamOutput(stderr)
	}()

	// Wait for output to be read
	wg.Wait()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		a.appendOutput(fmt.Sprintf("\n[Command exited with error: %v]", err))
	} else {
		a.appendOutput("\n[Command completed successfully]")
	}
}

// streamOutput reads from a reader, strips ANSI codes, and appends to output area
func (a *App) streamOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		clean := stripANSI(line)
		if strings.TrimSpace(clean) != "" {
			a.appendOutput(clean)
		}
	}
}

// showMainScreen displays the main menu
func (a *App) showMainScreen() {
	// Header with settings button
	titleLabel := widget.NewLabelWithStyle("TDL Manager", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		a.showSettingsScreen()
	})

	header := container.NewBorder(nil, nil, titleLabel, settingsBtn)

	// Account selector
	accountLabel := widget.NewLabel("Account:")
	accountSelect := widget.NewSelect([]string{"default"}, func(value string) {
		a.namespace = value
	})
	accountSelect.SetSelected(a.namespace)

	accountRow := container.NewHBox(accountLabel, accountSelect)

	// Main action buttons - Row 1
	loginBtn := widget.NewButtonWithIcon("SESSION", theme.LoginIcon(), func() {
		a.showLoginScreen()
	})
	downloadBtn := widget.NewButtonWithIcon("DOWNLOAD", theme.DownloadIcon(), func() {
		a.showDownloadScreen()
	})
	uploadBtn := widget.NewButtonWithIcon("UPLOAD", theme.UploadIcon(), func() {
		a.showUploadScreen()
	})
	forwardBtn := widget.NewButtonWithIcon("FORWARD", theme.MailForwardIcon(), func() {
		a.showForwardScreen()
	})

	row1 := container.NewGridWithColumns(4, loginBtn, downloadBtn, uploadBtn, forwardBtn)

	// Main action buttons - Row 2
	backupBtn := widget.NewButtonWithIcon("BACKUP", theme.DocumentSaveIcon(), func() {
		a.showBackupScreen()
	})
	recoverBtn := widget.NewButtonWithIcon("RECOVER", theme.HistoryIcon(), func() {
		a.showRecoverScreen()
	})
	listChatsBtn := widget.NewButtonWithIcon("LIST CHATS", theme.ListIcon(), func() {
		go a.runCommand("chat", "ls")
	})
	versionBtn := widget.NewButtonWithIcon("VERSION", theme.InfoIcon(), func() {
		go a.runCommand("version")
	})

	row2 := container.NewGridWithColumns(4, backupBtn, recoverBtn, listChatsBtn, versionBtn)

	buttonContainer := container.NewVBox(row1, row2)

	// Output area
	outputLabel := widget.NewLabel("Output:")
	outputScroll := container.NewScroll(a.outputArea)
	outputScroll.SetMinSize(fyne.NewSize(650, 300))

	clearOutputBtn := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), func() {
		a.clearOutput()
	})

	outputHeader := container.NewBorder(nil, nil, outputLabel, clearOutputBtn)
	outputContainer := container.NewBorder(outputHeader, nil, nil, nil, outputScroll)

	// Main layout
	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		accountRow,
		widget.NewSeparator(),
		buttonContainer,
		widget.NewSeparator(),
		outputContainer,
	)

	a.mainWindow.SetContent(container.NewPadded(content))
}

// showLoginScreen displays the login/session screen
func (a *App) showLoginScreen() {
	titleLabel := widget.NewLabelWithStyle("Session", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Info text
	infoText := widget.NewLabel(
		"Login via TERMINAL first (GUI cannot handle interactive prompts):\n\n" +
		"  tdl login              # Auto-detect from Telegram Desktop\n" +
		"  tdl login -T qr        # QR code login\n" +
		"  tdl login -T code      # Phone + verification code\n\n" +
		"After logging in via terminal, click CHECK SESSION to verify.")
	infoText.Wrapping = fyne.TextWrapWord

	checkSessionBtn := widget.NewButtonWithIcon("CHECK SESSION", theme.ConfirmIcon(), func() {
		go func() {
			a.clearOutput()
			a.appendOutput("Checking if logged in...")
			a.appendOutput("")
			a.runCommand("chat", "ls")
		}()
	})

	backBtn := widget.NewButtonWithIcon("BACK", theme.NavigateBackIcon(), func() {
		a.showMainScreen()
	})

	buttonRow := container.NewHBox(layout.NewSpacer(), checkSessionBtn, backBtn, layout.NewSpacer())

	// Output area
	outputLabel := widget.NewLabel("Output:")
	outputScroll := container.NewScroll(a.outputArea)
	outputScroll.SetMinSize(fyne.NewSize(650, 280))

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		infoText,
		widget.NewSeparator(),
		buttonRow,
		widget.NewSeparator(),
		outputLabel,
		outputScroll,
	)

	a.mainWindow.SetContent(container.NewPadded(content))
}

// showDownloadScreen displays the download options
func (a *App) showDownloadScreen() {
	titleLabel := widget.NewLabelWithStyle("Download", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// URL input
	urlLabel := widget.NewLabel("Enter Telegram URLs (one per line):")
	urlEntry := widget.NewMultiLineEntry()
	urlEntry.SetPlaceHolder("https://t.me/channel/123\nhttps://t.me/channel/456")
	urlEntry.SetMinRowsVisible(5)

	// Directory picker
	dirLabel := widget.NewLabel("Save to:")
	homeDir, _ := os.UserHomeDir()
	dirEntry := widget.NewEntry()
	dirEntry.SetText(filepath.Join(homeDir, "Downloads"))

	browseBtn := widget.NewButtonWithIcon("Browse...", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				dirEntry.SetText(uri.Path())
			}
		}, a.mainWindow)
	})

	dirRow := container.NewBorder(nil, nil, dirLabel, browseBtn, dirEntry)

	// Action buttons
	startDownloadBtn := widget.NewButtonWithIcon("START DOWNLOAD", theme.DownloadIcon(), func() {
		go func() {
			urls := strings.Split(urlEntry.Text, "\n")
			args := []string{"dl"}
			for _, url := range urls {
				url = strings.TrimSpace(url)
				if url != "" {
					args = append(args, "-u", url)
				}
			}
			if dirEntry.Text != "" {
				args = append(args, "-d", dirEntry.Text)
			}
			if len(args) > 1 {
				a.runCommand(args...)
			} else {
				a.appendOutput("Error: Please enter at least one URL")
			}
		}()
	})

	backBtn := widget.NewButtonWithIcon("BACK", theme.NavigateBackIcon(), func() {
		a.showMainScreen()
	})

	buttonRow := container.NewHBox(layout.NewSpacer(), startDownloadBtn, backBtn, layout.NewSpacer())

	// Output area
	outputLabel := widget.NewLabel("Output:")
	outputScroll := container.NewScroll(a.outputArea)
	outputScroll.SetMinSize(fyne.NewSize(650, 200))

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		urlLabel,
		urlEntry,
		dirRow,
		widget.NewSeparator(),
		buttonRow,
		widget.NewSeparator(),
		outputLabel,
		outputScroll,
	)

	a.mainWindow.SetContent(container.NewPadded(content))
}

// showUploadScreen displays the upload options
func (a *App) showUploadScreen() {
	titleLabel := widget.NewLabelWithStyle("Upload", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Selected files list
	selectedFiles := []string{}
	filesLabel := widget.NewLabel("Selected Files:")
	filesList := widget.NewMultiLineEntry()
	filesList.SetPlaceHolder("No files selected")
	filesList.Disable()
	filesList.SetMinRowsVisible(4)

	selectFilesBtn := widget.NewButtonWithIcon("Select Files...", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				path := reader.URI().Path()
				selectedFiles = append(selectedFiles, path)
				filesList.SetText(strings.Join(selectedFiles, "\n"))
				reader.Close()
			}
		}, a.mainWindow)
	})

	clearFilesBtn := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), func() {
		selectedFiles = []string{}
		filesList.SetText("")
	})

	fileButtonRow := container.NewHBox(selectFilesBtn, clearFilesBtn)

	// Destination selection
	destLabel := widget.NewLabel("Destination:")
	var savedMessages = true
	chatEntry := widget.NewEntry()
	chatEntry.SetPlaceHolder("Chat username or ID")
	chatEntry.Disable()

	destRadio := widget.NewRadioGroup([]string{
		"Saved Messages",
		"Chat:",
	}, func(value string) {
		savedMessages = (value == "Saved Messages")
		if savedMessages {
			chatEntry.Disable()
		} else {
			chatEntry.Enable()
		}
	})
	destRadio.SetSelected("Saved Messages")

	destRow := container.NewHBox(destRadio, chatEntry)

	// Action buttons
	startUploadBtn := widget.NewButtonWithIcon("START UPLOAD", theme.UploadIcon(), func() {
		go func() {
			if len(selectedFiles) == 0 {
				a.appendOutput("Error: Please select at least one file")
				return
			}

			args := []string{"up"}
			for _, file := range selectedFiles {
				args = append(args, "-p", file)
			}
			if !savedMessages && chatEntry.Text != "" {
				args = append(args, "-c", chatEntry.Text)
			}
			a.runCommand(args...)
		}()
	})

	backBtn := widget.NewButtonWithIcon("BACK", theme.NavigateBackIcon(), func() {
		a.showMainScreen()
	})

	buttonRow := container.NewHBox(layout.NewSpacer(), startUploadBtn, backBtn, layout.NewSpacer())

	// Output area
	outputLabel := widget.NewLabel("Output:")
	outputScroll := container.NewScroll(a.outputArea)
	outputScroll.SetMinSize(fyne.NewSize(650, 180))

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		filesLabel,
		filesList,
		fileButtonRow,
		widget.NewSeparator(),
		destLabel,
		destRow,
		widget.NewSeparator(),
		buttonRow,
		widget.NewSeparator(),
		outputLabel,
		outputScroll,
	)

	a.mainWindow.SetContent(container.NewPadded(content))
}

// showForwardScreen displays the forward options
func (a *App) showForwardScreen() {
	titleLabel := widget.NewLabelWithStyle("Forward Messages", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Source URL
	fromLabel := widget.NewLabel("From URL:")
	fromEntry := widget.NewEntry()
	fromEntry.SetPlaceHolder("https://t.me/channel/123")

	// Destination selection
	destLabel := widget.NewLabel("Destination:")
	var toSavedMessages = true
	toEntry := widget.NewEntry()
	toEntry.SetPlaceHolder("Chat username or ID")
	toEntry.Disable()

	destRadio := widget.NewRadioGroup([]string{
		"Saved Messages",
		"Chat:",
	}, func(value string) {
		toSavedMessages = (value == "Saved Messages")
		if toSavedMessages {
			toEntry.Disable()
		} else {
			toEntry.Enable()
		}
	})
	destRadio.SetSelected("Saved Messages")

	destRow := container.NewHBox(destRadio, toEntry)

	// Action buttons
	startForwardBtn := widget.NewButtonWithIcon("START FORWARD", theme.MailForwardIcon(), func() {
		go func() {
			if fromEntry.Text == "" {
				a.appendOutput("Error: Please enter a source URL")
				return
			}

			args := []string{"forward", "--from", fromEntry.Text}
			if !toSavedMessages && toEntry.Text != "" {
				args = append(args, "--to", toEntry.Text)
			}
			a.runCommand(args...)
		}()
	})

	backBtn := widget.NewButtonWithIcon("BACK", theme.NavigateBackIcon(), func() {
		a.showMainScreen()
	})

	buttonRow := container.NewHBox(layout.NewSpacer(), startForwardBtn, backBtn, layout.NewSpacer())

	// Output area
	outputLabel := widget.NewLabel("Output:")
	outputScroll := container.NewScroll(a.outputArea)
	outputScroll.SetMinSize(fyne.NewSize(650, 200))

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		fromLabel,
		fromEntry,
		widget.NewSeparator(),
		destLabel,
		destRow,
		widget.NewSeparator(),
		buttonRow,
		widget.NewSeparator(),
		outputLabel,
		outputScroll,
	)

	a.mainWindow.SetContent(container.NewPadded(content))
}

// showBackupScreen displays the backup options
func (a *App) showBackupScreen() {
	titleLabel := widget.NewLabelWithStyle("Backup Account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Backup destination
	destLabel := widget.NewLabel("Backup destination (optional):")
	homeDir, _ := os.UserHomeDir()
	destEntry := widget.NewEntry()
	destEntry.SetPlaceHolder(filepath.Join(homeDir, "backup.tdl"))

	browseBtn := widget.NewButtonWithIcon("Browse...", theme.FolderOpenIcon(), func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				destEntry.SetText(writer.URI().Path())
				writer.Close()
			}
		}, a.mainWindow)
	})

	destRow := container.NewBorder(nil, nil, destLabel, browseBtn, destEntry)

	// Action buttons
	startBackupBtn := widget.NewButtonWithIcon("START BACKUP", theme.DocumentSaveIcon(), func() {
		go func() {
			args := []string{"backup"}
			if destEntry.Text != "" {
				args = append(args, "-d", destEntry.Text)
			}
			a.runCommand(args...)
		}()
	})

	backBtn := widget.NewButtonWithIcon("BACK", theme.NavigateBackIcon(), func() {
		a.showMainScreen()
	})

	buttonRow := container.NewHBox(layout.NewSpacer(), startBackupBtn, backBtn, layout.NewSpacer())

	// Output area
	outputLabel := widget.NewLabel("Output:")
	outputScroll := container.NewScroll(a.outputArea)
	outputScroll.SetMinSize(fyne.NewSize(650, 250))

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		destRow,
		widget.NewSeparator(),
		buttonRow,
		widget.NewSeparator(),
		outputLabel,
		outputScroll,
	)

	a.mainWindow.SetContent(container.NewPadded(content))
}

// showRecoverScreen displays the recover options
func (a *App) showRecoverScreen() {
	titleLabel := widget.NewLabelWithStyle("Recover Account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Backup file selection
	fileLabel := widget.NewLabel("Backup file:")
	fileEntry := widget.NewEntry()
	fileEntry.SetPlaceHolder("Select backup file...")

	browseBtn := widget.NewButtonWithIcon("Browse...", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				fileEntry.SetText(reader.URI().Path())
				reader.Close()
			}
		}, a.mainWindow)
	})

	fileRow := container.NewBorder(nil, nil, fileLabel, browseBtn, fileEntry)

	// Action buttons
	startRecoverBtn := widget.NewButtonWithIcon("START RECOVER", theme.HistoryIcon(), func() {
		go func() {
			if fileEntry.Text == "" {
				a.appendOutput("Error: Please select a backup file")
				return
			}
			a.runCommand("recover", "-f", fileEntry.Text)
		}()
	})

	backBtn := widget.NewButtonWithIcon("BACK", theme.NavigateBackIcon(), func() {
		a.showMainScreen()
	})

	buttonRow := container.NewHBox(layout.NewSpacer(), startRecoverBtn, backBtn, layout.NewSpacer())

	// Output area
	outputLabel := widget.NewLabel("Output:")
	outputScroll := container.NewScroll(a.outputArea)
	outputScroll.SetMinSize(fyne.NewSize(650, 250))

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		fileRow,
		widget.NewSeparator(),
		buttonRow,
		widget.NewSeparator(),
		outputLabel,
		outputScroll,
	)

	a.mainWindow.SetContent(container.NewPadded(content))
}

// showSettingsScreen displays the settings options
func (a *App) showSettingsScreen() {
	titleLabel := widget.NewLabelWithStyle("Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Account name
	accountLabel := widget.NewLabel("Account Name:")
	accountEntry := widget.NewEntry()
	accountEntry.SetText(a.namespace)
	accountEntry.SetPlaceHolder("default")

	// Proxy
	proxyLabel := widget.NewLabel("Proxy:")
	proxyEntry := widget.NewEntry()
	proxyEntry.SetText(a.proxy)
	proxyEntry.SetPlaceHolder("socks5://localhost:1080")

	// Form layout
	form := container.NewVBox(
		accountLabel,
		accountEntry,
		proxyLabel,
		proxyEntry,
	)

	// Action buttons
	saveBtn := widget.NewButtonWithIcon("SAVE", theme.DocumentSaveIcon(), func() {
		a.namespace = accountEntry.Text
		if a.namespace == "" {
			a.namespace = "default"
		}
		a.proxy = proxyEntry.Text
		a.showMainScreen()
	})

	backBtn := widget.NewButtonWithIcon("BACK", theme.NavigateBackIcon(), func() {
		a.showMainScreen()
	})

	buttonRow := container.NewHBox(layout.NewSpacer(), saveBtn, backBtn, layout.NewSpacer())

	// Output area (for testing connection, etc.)
	outputLabel := widget.NewLabel("Output:")
	outputScroll := container.NewScroll(a.outputArea)
	outputScroll.SetMinSize(fyne.NewSize(650, 200))

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		buttonRow,
		widget.NewSeparator(),
		outputLabel,
		outputScroll,
	)

	a.mainWindow.SetContent(container.NewPadded(content))
}



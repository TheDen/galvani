package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/caseymrm/menuet"
)

const boltIconOutline = "bolt.png"
const boltIconFilled = "bolt-filled.png"

func runCommand(str string) error {
	//	script := fmt.Sprintf("set the clipboard to \"%s\"", str)
	cmd := exec.Command("/usr/bin/osascript", "-e", fmt.Sprintf("do shell script \"%s\" with prompt \"Galvani is trying to update battery prefrences\" with administrator privileges", str))
	err := cmd.Run()
	return err
}

func setMenuStatesFalse() {
	menuet.Defaults().SetBoolean("alwaysState", false)
	menuet.Defaults().SetBoolean("neverState", false)
	menuet.Defaults().SetBoolean("batteryOnlyState", false)
	menuet.Defaults().SetBoolean("powerOnlyState", false)
}

func setIconState() string {
	cmd := exec.Command("pmset", "-g")
	output, err := cmd.Output()
	if err != nil {
		return boltIconOutline
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "lowpowermode") {
			fields := strings.Fields(line)
			if fields[1] == "1" {
				return boltIconFilled
			}
			return boltIconOutline
		}
	}
	return boltIconOutline
}

// osascript -e 'do shell script "sudo pmset -a lowpowermode 1" with administrator privileges'
func menuItems() []menuet.MenuItem {
	alwaysState := menuet.Defaults().Boolean("alwaysState")
	neverState := menuet.Defaults().Boolean("neverState")
	batteryOnlyState := menuet.Defaults().Boolean("batteryOnlyState")
	powerOnlyState := menuet.Defaults().Boolean("powerOnlyState")

	items := []menuet.MenuItem{}
	items = append(items, menuet.MenuItem{
		Text:     "Low Power Mode",
		FontSize: 15,
	})

	items = append(items, menuet.MenuItem{
		Text: fmt.Sprintf("💡 Always"),
		Clicked: func() {
			runCommand("sudo pmset -a lowpowermode 1")
			setMenuStatesFalse()
			menuet.Defaults().SetBoolean("alwaysState", true)
		},
		State: alwaysState,
	})

	items = append(items, menuet.MenuItem{
		Text: fmt.Sprintf("🛑 Never"),
		Clicked: func() {
			runCommand("sudo pmset -a lowpowermode 0")
			setMenuStatesFalse()
			menuet.Defaults().SetBoolean("neverState", true)
		},
		State: neverState,
	})

	items = append(items, menuet.MenuItem{
		Text: fmt.Sprintf("🔋 Only on Battery"),
		Clicked: func() {
			runCommand("sudo pmset -a lowpowermode 0; sudo pmset -b lowpowermode 1")
			setMenuStatesFalse()
			menuet.Defaults().SetBoolean("batteryOnlyState", true)
		},
		State: batteryOnlyState,
	})

	items = append(items, menuet.MenuItem{
		Text: fmt.Sprintf("🔌 Only on Power"),
		Clicked: func() {
			runCommand("sudo pmset -a lowpowermode 0; sudo pmset -c lowpowermode 1")
			menuet.Defaults().SetBoolean("alwaysState", false)
			menuet.Defaults().SetBoolean("neverState", false)
			menuet.Defaults().SetBoolean("batteryOnlyState", false)
			menuet.Defaults().SetBoolean("powerOnlyState", true)
		},
		State: powerOnlyState,
	})

	return items
}

func menu() {
	for {
		menuet.App().SetMenuState(&menuet.MenuState{
			Image: setIconState(),
		})
		menuet.App().MenuChanged()
		time.Sleep(time.Second)
	}
}

func main() {
	go menu()
	app := menuet.App()
	app.Name = "Galvani"
	app.Label = "com.github.theden.galvani"
	app.Children = menuItems
	app.AutoUpdate.Version = "0.0.2"
	app.AutoUpdate.Repo = "TheDen/galvani"
	app.RunApplication()
}

package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/caseymrm/menuet"
)

func runCommand(str string) error {
	//	script := fmt.Sprintf("set the clipboard to \"%s\"", str)
	cmd := exec.Command("/usr/bin/osascript", "-e", fmt.Sprintf("do shell script \"%s\" with prompt \"Galvani is trying to update battery prefrences\" with administrator privileges", str))
	err := cmd.Run()
	return err
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
			menuet.Defaults().SetBoolean("alwaysState", true)
			menuet.Defaults().SetBoolean("neverState", false)
			menuet.Defaults().SetBoolean("batteryOnlyState", false)
			menuet.Defaults().SetBoolean("powerOnlyState", false)
		},
		State: alwaysState,
	})

	items = append(items, menuet.MenuItem{
		Text: fmt.Sprintf("🛑 Never"),
		Clicked: func() {
			runCommand("sudo pmset -a lowpowermode 0")
			menuet.Defaults().SetBoolean("alwaysState", false)
			menuet.Defaults().SetBoolean("neverState", true)
			menuet.Defaults().SetBoolean("batteryOnlyState", false)
			menuet.Defaults().SetBoolean("powerOnlyState", false)
		},
		State: neverState,
	})

	items = append(items, menuet.MenuItem{
		Text: fmt.Sprintf("🔋 Only on Battery"),
		Clicked: func() {
			runCommand("sudo pmset -a lowpowermode 0; sudo pmset -b lowpowermode 1")
			menuet.Defaults().SetBoolean("alwaysState", false)
			menuet.Defaults().SetBoolean("neverState", false)
			menuet.Defaults().SetBoolean("batteryOnlyState", true)
			menuet.Defaults().SetBoolean("powerOnlyState", false)
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
			Image: "bolt.png",
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
	app.RunApplication()
}

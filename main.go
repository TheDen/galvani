package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/caseymrm/menuet"
	"howett.net/plist"
)

const appVersion = "0.0.6"
const boltIconOutline = "bolt.png"
const boltIconFilled = "bolt-filled.png"

var lowPowerMode = ""

func runCommand(str string) error {
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

func checkStateAtStartup() {
	hardwareUUID, err := getHardwareUUID()
	if err != nil {
		return
	}
	plistPath := fmt.Sprintf("/Library/Preferences/com.apple.PowerManagement.%s.plist", hardwareUUID)
	acLowPowerMode, batteryLowPowerMode, err := getPowerState(plistPath)
	if err != nil {
		return
	}
	if acLowPowerMode == 1 && batteryLowPowerMode == 1 {
		setMenuStatesFalse()
		menuet.Defaults().SetBoolean("alwaysState", true)
	} else if acLowPowerMode == 0 && batteryLowPowerMode == 0 {
		setMenuStatesFalse()
		menuet.Defaults().SetBoolean("neverState", true)
	} else if acLowPowerMode == 0 && batteryLowPowerMode == 1 {
		setMenuStatesFalse()
		menuet.Defaults().SetBoolean("batteryOnlyState", true)
	} else if acLowPowerMode == 1 && batteryLowPowerMode == 0 {
		setMenuStatesFalse()
		menuet.Defaults().SetBoolean("powerOnlyState", true)
	}
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
				lowPowerMode = "ON"
				return boltIconFilled
			}
			lowPowerMode = "OFF"
			return boltIconOutline
		}
	}
	return boltIconOutline
}

func menuItems() []menuet.MenuItem {
	alwaysState := menuet.Defaults().Boolean("alwaysState")
	neverState := menuet.Defaults().Boolean("neverState")
	batteryOnlyState := menuet.Defaults().Boolean("batteryOnlyState")
	powerOnlyState := menuet.Defaults().Boolean("powerOnlyState")

	items := []menuet.MenuItem{}
	items = append(items, menuet.MenuItem{
		Text:     fmt.Sprintf("Galvani (v%s)", appVersion),
		FontSize: 12,
	})

	items = append(items, menuet.MenuItem{
		Text:     fmt.Sprintf("Low Power Mode %s", lowPowerMode),
		FontSize: 12,
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
			setMenuStatesFalse()
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

func getHardwareUUID() (string, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Hardware UUID:") {
			uuid := strings.TrimSpace(strings.Split(line, ":")[1])
			return uuid, nil
		}
	}
	return "", fmt.Errorf("Hardware UUID not found")
}

func getPowerState(plistPath string) (uint64, uint64, error) {
	// open the plist file
	f, err := os.Open(plistPath)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	// create a map to store the unmarshaled data
	var m map[string]interface{}

	// decode the plist file into the map
	decoder := plist.NewDecoder(f)
	err = decoder.Decode(&m)
	if err != nil {
		return 0, 0, err
	}

	// extract the LowPowerMode values for Battery and AC
	batteryLowPowerMode := m["Battery Power"].(map[string]interface{})["LowPowerMode"].(uint64)
	acLowPowerMode := m["AC Power"].(map[string]interface{})["LowPowerMode"].(uint64)

	return acLowPowerMode, batteryLowPowerMode, nil
}

func main() {
	go menu()
	checkStateAtStartup()
	app := menuet.App()
	app.Name = "Galvani"
	app.Label = "com.github.theden.galvani"
	app.Children = menuItems
	app.AutoUpdate.Version = appVersion
	app.AutoUpdate.Repo = "TheDen/galvani"
	app.RunApplication()
}

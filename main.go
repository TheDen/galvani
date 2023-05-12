package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/caseymrm/menuet"
	"howett.net/plist"
)

const (
	appVersion      = "0.0.9"
	boltIconOutline = "bolt.png"
	boltIconFilled  = "bolt-filled.png"
)

var lowPowerMode = ""

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

func setLowPowerMode(str string) error {
	cmd := exec.Command("/usr/bin/osascript", "-e", fmt.Sprintf("do shell script \"%s\" with prompt \"Galvani is trying to update battery prefrences\" with administrator privileges", str))
	err := cmd.Run()
	return err
}

func checkLowPowerState(hardwareUUID string) {
	fmt.Println(hardwareUUID)
	plistPath := fmt.Sprintf("/Library/Preferences/com.apple.PowerManagement.%s.plist", hardwareUUID)

	for {
		cmd := exec.Command("defaults", "read", plistPath)
		out, err := cmd.Output()
		if err != nil {
			fmt.Println("Command execution error:", err)
			continue
		}

		var config map[string]interface{}
		_, err = plist.Unmarshal(out, &config)
		if err != nil {
			fmt.Println("Failed to decode plist:", err)
			continue
		}

		// extract the LowPowerMode values for Battery and AC
		batteryLowPowerModeStr := config["Battery Power"].(map[string]interface{})["LowPowerMode"].(string)
		batteryLowPowerMode, err := strconv.ParseUint(batteryLowPowerModeStr, 10, 64)
		if err != nil {
			fmt.Println("Failed to parse Battery Power LowPowerMode:", err)
			continue
		}

		acLowPowerModeStr := config["AC Power"].(map[string]interface{})["LowPowerMode"].(string)
		acLowPowerMode, err := strconv.ParseUint(acLowPowerModeStr, 10, 64)
		if err != nil {
			fmt.Println("Failed to parse AC Power LowPowerMode:", err)
			continue
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

		time.Sleep(time.Second)
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

func setMenuStatesFalse() {
	menuet.Defaults().SetBoolean("alwaysState", false)
	menuet.Defaults().SetBoolean("neverState", false)
	menuet.Defaults().SetBoolean("batteryOnlyState", false)
	menuet.Defaults().SetBoolean("powerOnlyState", false)
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
			err := setLowPowerMode("sudo pmset -a lowpowermode 1")
			if err == nil {
				setMenuStatesFalse()
				menuet.Defaults().SetBoolean("alwaysState", true)
			}
		},
		State: alwaysState,
	})

	items = append(items, menuet.MenuItem{
		Text: fmt.Sprintf("🛑 Never"),
		Clicked: func() {
			err := setLowPowerMode("sudo pmset -a lowpowermode 0")
			if err == nil {
				setMenuStatesFalse()
				menuet.Defaults().SetBoolean("neverState", true)
			}
		},
		State: neverState,
	})

	items = append(items, menuet.MenuItem{
		Text: fmt.Sprintf("🔋 Only on Battery"),
		Clicked: func() {
			err := setLowPowerMode("sudo pmset -c lowpowermode 0; sudo pmset -b lowpowermode 1")
			if err == nil {
				setMenuStatesFalse()
				menuet.Defaults().SetBoolean("batteryOnlyState", true)
			}
		},
		State: batteryOnlyState,
	})

	items = append(items, menuet.MenuItem{
		Text: fmt.Sprintf("🔌 Only on Power"),
		Clicked: func() {
			err := setLowPowerMode("sudo pmset -c lowpowermode 1; sudo pmset -b lowpowermode 0")
			if err == nil {
				setMenuStatesFalse()
				menuet.Defaults().SetBoolean("powerOnlyState", true)
			}
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
	hardwareUUID, err := getHardwareUUID()
	if err == nil {
		go checkLowPowerState(hardwareUUID)
	}

	app := menuet.App()
	app.Name = "Galvani"
	app.Label = "com.github.theden.galvani"
	app.Children = menuItems
	app.AutoUpdate.Version = appVersion
	app.AutoUpdate.Repo = "TheDen/galvani"
	app.RunApplication()
}

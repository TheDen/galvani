package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/caseymrm/menuet"
	"howett.net/plist"
)

const (
	appVersion      = "0.1.1"
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

func updateLowPowerStateMenu(hardwareUUID string) {
	log.Printf("Hardware UUID is %s\n", hardwareUUID)
	plistPath := fmt.Sprintf("/Library/Preferences/com.apple.PowerManagement.%s.plist", hardwareUUID)
	currentState := ""

	for {
		cmd := exec.Command("defaults", "read", plistPath)
		out, err := cmd.Output()
		if err != nil {
			log.Println(err)
			continue
		}

		var config map[string]interface{}
		_, err = plist.Unmarshal(out, &config)
		if err != nil {
			log.Println(err)
			continue
		}

		// extract the LowPowerMode values for Battery and AC
		batteryLowPowerModeStr := config["Battery Power"].(map[string]interface{})["LowPowerMode"].(string)
		batteryLowPowerMode, err := strconv.ParseUint(batteryLowPowerModeStr, 10, 64)
		if err != nil {
			log.Println(err)
			continue
		}

		acLowPowerModeStr := config["AC Power"].(map[string]interface{})["LowPowerMode"].(string)
		acLowPowerMode, err := strconv.ParseUint(acLowPowerModeStr, 10, 64)
		if err != nil {
			log.Println(err)
			continue
		}

		states := map[[2]uint64]string{
			{1, 1}: "alwaysState",
			{0, 0}: "neverState",
			{0, 1}: "batteryOnlyState",
			{1, 0}: "powerOnlyState",
		}

		// Get the state for the current condition
		state, _ := states[[2]uint64{acLowPowerMode, batteryLowPowerMode}]
		// Only update if state has changed
		if state != currentState {
			setMenuStatesFalse()
			menuet.Defaults().SetBoolean(state, true)
			log.Printf("Updated state from %s to %s\n", currentState, state)
			currentState = state
		}
		time.Sleep(time.Second)
	}
}

func updateCurrentState(currentIconState string) string {
	cmd := exec.Command("pmset", "-g")
	output, err := cmd.Output()
	if err != nil {
		log.Println(err)
		return currentIconState
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
	return currentIconState
}

func setMenuStatesFalse() {
	keys := []string{"alwaysState", "neverState", "batteryOnlyState", "powerOnlyState"}
	for _, key := range keys {
		menuet.Defaults().SetBoolean(key, false)
	}
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
	currentIconState := ""
	newIconState := ""
	for {
		newIconState = updateCurrentState(currentIconState)
		if currentIconState != newIconState {
			menuet.App().SetMenuState(&menuet.MenuState{
				Image: newIconState,
			})
			menuet.App().MenuChanged()
			currentIconState = newIconState
		}
		time.Sleep(time.Second)
	}
}

func main() {
	go menu()
	hardwareUUID, err := getHardwareUUID()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	go updateLowPowerStateMenu(hardwareUUID)

	app := menuet.App()
	app.Name = "Galvani"
	app.Label = "com.github.theden.galvani"
	app.Children = menuItems
	app.AutoUpdate.Version = appVersion
	app.AutoUpdate.Repo = "TheDen/galvani"
	app.RunApplication()
}

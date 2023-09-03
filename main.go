package main

import (
	"errors"
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
	appVersion      = "0.1.3"
	boltIconOutline = "bolt.png"
	boltIconFilled  = "bolt-filled.png"
)

type BatteryState uint8

const (
	ALWAYS BatteryState = iota
	NEVER
	BATTERY_ONLY
	POWER_ONLY
)

func (b BatteryState) String() string {
	switch b {
	case ALWAYS:
		return "alwaysState"
	case NEVER:
		return "neverState"
	case BATTERY_ONLY:
		return "batteryOnlyState"
	case POWER_ONLY:
		return "powerOnlyState"
	default:
		return "Invalid state"
	}
}

func getBatteryStates() []BatteryState {
	return []BatteryState{ALWAYS, NEVER, BATTERY_ONLY, POWER_ONLY}
}

func getStateFromCondition(ac bool, battery bool) BatteryState {
	states := map[[2]bool]BatteryState{
		{true, true}:   ALWAYS,
		{false, false}: NEVER,
		{false, true}:  BATTERY_ONLY,
		{true, false}:  POWER_ONLY,
	}
	return BatteryState(states[[2]bool{ac, battery}])
}

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
	return "", errors.New("hardware UUID not found")
}

func setLowPowerMode(str string) error {
	cmd := exec.Command(
		"/usr/bin/osascript",
		"-e",
		fmt.Sprintf(
			"do shell script \"%s\" with prompt \"Galvani is trying to update battery prefrences\" with administrator privileges",
			str,
		),
	)
	err := cmd.Run()
	return err
}

func updateLowPowerStateMenu(hardwareUUID string) {
	log.Printf("Hardware UUID is %s\n", hardwareUUID)
	plistPath := fmt.Sprintf(
		"/Library/Preferences/com.apple.PowerManagement.%s.plist",
		hardwareUUID,
	)
	var currentState BatteryState
	tick := time.Tick(1 * time.Second)

	for range tick {
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
		batteryLowPowerMode, err := strconv.ParseBool(batteryLowPowerModeStr)
		if err != nil {
			log.Println(err)
			continue
		}

		acLowPowerModeStr := config["AC Power"].(map[string]interface{})["LowPowerMode"].(string)
		acLowPowerMode, err := strconv.ParseBool(acLowPowerModeStr)
		if err != nil {
			log.Println(err)
			continue
		}

		// Get the state for the current condition
		state := getStateFromCondition(acLowPowerMode, batteryLowPowerMode)
		// Only update if state has changed
		if state != currentState {
			setMenuStatesFalse()
			menuet.Defaults().SetBoolean(state.String(), true)
			log.Printf("Updated state from %s to %s\n", currentState, state)
			currentState = state
		}
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
	for _, key := range getBatteryStates() {
		menuet.Defaults().SetBoolean(key.String(), false)
	}
}

func menuItems() []menuet.MenuItem {
	alwaysState := menuet.Defaults().Boolean(ALWAYS.String())
	neverState := menuet.Defaults().Boolean(NEVER.String())
	batteryOnlyState := menuet.Defaults().Boolean(BATTERY_ONLY.String())
	powerOnlyState := menuet.Defaults().Boolean(POWER_ONLY.String())

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
		Text: "💡 Always",
		Clicked: func() {
			err := setLowPowerMode("sudo pmset -a lowpowermode 1")
			if err == nil {
				setMenuStatesFalse()
				menuet.Defaults().SetBoolean(ALWAYS.String(), true)
			}
		},
		State: alwaysState,
	})

	items = append(items, menuet.MenuItem{
		Text: "🛑 Never",
		Clicked: func() {
			err := setLowPowerMode("sudo pmset -a lowpowermode 0")
			if err == nil {
				setMenuStatesFalse()
				menuet.Defaults().SetBoolean(NEVER.String(), true)
			}
		},
		State: neverState,
	})

	items = append(items, menuet.MenuItem{
		Text: "🔋 Only on Battery",
		Clicked: func() {
			err := setLowPowerMode("sudo pmset -c lowpowermode 0; sudo pmset -b lowpowermode 1")
			if err == nil {
				setMenuStatesFalse()
				menuet.Defaults().SetBoolean(BATTERY_ONLY.String(), true)
			}
		},
		State: batteryOnlyState,
	})

	items = append(items, menuet.MenuItem{
		Text: "🔌 Only on Power",
		Clicked: func() {
			err := setLowPowerMode("sudo pmset -c lowpowermode 1; sudo pmset -b lowpowermode 0")
			if err == nil {
				setMenuStatesFalse()
				menuet.Defaults().SetBoolean(POWER_ONLY.String(), true)
			}
		},
		State: powerOnlyState,
	})

	return items
}

func menu() {
	currentIconState := ""
	newIconState := ""
	tick := time.Tick(1 * time.Second)
	for range tick {
		newIconState = updateCurrentState(currentIconState)
		if currentIconState != newIconState {
			menuet.App().SetMenuState(&menuet.MenuState{
				Image: newIconState,
			})
			menuet.App().MenuChanged()
			currentIconState = newIconState
		}
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

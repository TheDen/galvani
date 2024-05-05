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
	appVersion                   = "0.2.4"
	boltIconOutline              = "bolt.png"
	boltIconFilled               = "bolt-filled.png"
	ALWAYS          BatteryState = iota
	NEVER
	BATTERY_ONLY
	POWER_ONLY
)

var (
	hardwareUUID string //lint:ignore U1000 false positives
	plistPath    string
	lowPowerMode = ""
	inChan       = make(chan BatteryState)
	currentState BatteryState //lint:ignore U1000 false positives
	currentIcon  string
)

type BatteryState uint8

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
			"do shell script \"%s\" with prompt \"Galvani is trying to update battery preferences\" with administrator privileges",
			str,
		),
	)
	err := cmd.Run()
	return err
}

func getState() (BatteryState, error) {
	cmd := exec.Command("defaults", "read", plistPath)
	out, err := cmd.Output()
	if err != nil {
		return NEVER, err
	}

	var config map[string]interface{}
	_, err = plist.Unmarshal(out, &config)
	if err != nil {
		return NEVER, err
	}

	// extract the LowPowerMode values for Battery and AC
	batteryLowPowerModeStr := config["Battery Power"].(map[string]interface{})["LowPowerMode"].(string)
	batteryLowPowerMode, err := strconv.ParseBool(batteryLowPowerModeStr)
	if err != nil {
		return NEVER, err
	}

	acLowPowerModeStr := config["AC Power"].(map[string]interface{})["LowPowerMode"].(string)
	acLowPowerMode, err := strconv.ParseBool(acLowPowerModeStr)
	if err != nil {
		return NEVER, err
	}

	// Get the state for the current condition
	state := getStateFromCondition(acLowPowerMode, batteryLowPowerMode)
	return state, nil
}

func pollLowPowerState() {
	// Set initial state
	currentState, err := getState()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = setMenu(currentState)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// Polling loop
	tick := time.Tick(15 * time.Second)
	for range tick {
		state, err := getState()
		if err != nil {
			log.Println(err)
			continue
		}

		// Only update if state has changed
		if state != currentState {
			inChan <- state
			log.Printf("Updated state from %s to %s\n", currentState, state)
			currentState = state
			continue
		}

		icon, err := getIcon()
		if err != nil {
			log.Println(err)
			continue
		}
		if icon != currentIcon {
			setIcon(icon)
		}
	}
}

func getIcon() (string, error) {
	cmd := exec.Command("pmset", "-g")
	output, err := cmd.Output()
	if err != nil {
		log.Println(err)
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "lowpowermode") {
			fields := strings.Fields(line)
			if fields[1] == "1" {
				lowPowerMode = "ON"
				return boltIconFilled, nil
			}
			lowPowerMode = "OFF"
			return boltIconOutline, nil
		}
	}
	return "", errors.New("low power mode not found")
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
				inChan <- ALWAYS
			}
		},
		State: alwaysState,
	})

	items = append(items, menuet.MenuItem{
		Text: "🛑 Never",
		Clicked: func() {
			err := setLowPowerMode("sudo pmset -a lowpowermode 0")
			if err == nil {
				inChan <- NEVER
			}
		},
		State: neverState,
	})

	items = append(items, menuet.MenuItem{
		Text: "🔋 Only on Battery",
		Clicked: func() {
			err := setLowPowerMode("sudo pmset -c lowpowermode 0; sudo pmset -b lowpowermode 1")
			if err == nil {
				inChan <- BATTERY_ONLY
			}
		},
		State: batteryOnlyState,
	})

	items = append(items, menuet.MenuItem{
		Text: "🔌 Only on Power",
		Clicked: func() {
			err := setLowPowerMode("sudo pmset -c lowpowermode 1; sudo pmset -b lowpowermode 0")
			if err == nil {
				inChan <- POWER_ONLY
			}
		},
		State: powerOnlyState,
	})

	return items
}

func setIcon(icon string) {
	menuet.App().SetMenuState(&menuet.MenuState{
		Image: icon,
	})
	menuet.App().MenuChanged()
	currentIcon = icon
}

func setMenu(state BatteryState) error {
	setMenuStatesFalse()
	menuet.Defaults().SetBoolean(state.String(), true)
	icon, err := getIcon()
	if err != nil {
		return err
	}
	setIcon(icon)
	currentState = state
	return nil
}

func menu() {
	for state := range inChan {
		err := setMenu(state)
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	go menu()
	hardwareUUID, err := getHardwareUUID()
	log.Printf("Hardware UUID is %s\n", hardwareUUID)
	plistPath = fmt.Sprintf(
		"/Library/Preferences/com.apple.PowerManagement.%s.plist",
		hardwareUUID,
	)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	go pollLowPowerState()

	app := menuet.App()
	app.Name = "Galvani"
	app.Label = "com.github.theden.galvani"
	app.Children = menuItems
	app.AutoUpdate.Version = appVersion
	app.AutoUpdate.Repo = "TheDen/galvani"
	app.RunApplication()
}

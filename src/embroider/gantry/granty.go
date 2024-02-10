package gantry

import (
	"embroider/gcode"
	"embroider/serial"
	"fmt"
	"log"
	"strings"
	"time"
)

type Gantry struct {
	port serial.Serial
}

func (g *Gantry) disableLock() {
	if g.port.HandleCommand([]serial.Transmission{
		serial.Transmission{
			Out: []string{},
			In:  []string{"['$H'|'$X' to unlock]"},
		},
	}, 3) {
		var disable_selection string
		for disable_selection != "Y" && disable_selection != "n" {
			fmt.Printf("Lock is enabled. Disable the lock? [Y/n]\n")
			fmt.Printf("> ")
			fmt.Scanln(&disable_selection)
		}
		if disable_selection == "Y" {
			for !g.port.HandleCommand([]serial.Transmission{
				serial.Transmission{
					Out: []string{"$X"},
					In:  []string{"[Caution: Unlocked]", "ok", "ok"},
				},
			}, 3) {

			}
		} else {
			log.Fatal("Exiting.")
		}
	}
}

type status struct {
	running bool
	mpos    [3]float32
	wpos    [3]float32
}

func (g *Gantry) getStatus() (status, bool) {
	success, incoming := g.port.HandleRequest([]string{"?"}, 3, 3)
	if !success {
		return status{}, false
	}

	var result status
	if (len(incoming) == 0) {
		return status{}, false
	}

	current_status := incoming[0]
	current_status = strings.TrimPrefix(current_status, "<")
	current_status = strings.TrimSuffix(current_status, ">")

	state_idx := strings.Index(current_status, ",")
	if (state_idx == -1) {
		return status{}, false
	}
	state := current_status[:state_idx]
	if (len(current_status) <= state_idx + 1) {
		return status{}, false
	}
	current_status = current_status[state_idx+1:]

	n, err := fmt.Sscanf(current_status, "MPos:%g,%g,%g,WPos:%g,%g,%g",
		&result.mpos[0], &result.mpos[1], &result.mpos[2],
		&result.wpos[0], &result.wpos[1], &result.wpos[2])

	if ((n != 6) || (err != nil)) {
		return status{}, false
	}

	if state == "Run" {
		result.running = true
	}

	return result, true
}

func (g *Gantry) HandleGcode(code gcode.GCode, retry int) bool {
	for {
		if !g.port.HandleCommand([]serial.Transmission{
			serial.Transmission{
				Out: []string{code.ToString()},
				In:  []string{"ok", "ok"},
			},
		}, 3) {
			fmt.Printf("Command \"%s\" failed...\n", code.ToString())
			continue
		}

		status, success := g.getStatus()
		for status.running && success {
			time.Sleep(250 * time.Millisecond)
			status, success = g.getStatus()
			fmt.Printf("Position: %.2f %.2f\n", status.mpos[0], status.mpos[1])
		}
		if !success {
			fmt.Printf("Failed to get status.\n")
			continue
		}

		fmt.Printf("Successfully completed gantry operation\n")
		return true
	}

	return false
}

func Create(port string) Gantry {
	var gantry Gantry

	serial_success := false
	for !serial_success {
		gantry.port, serial_success = serial.SetupSerialCommunication(port, []string{}, []string{"Grbl 0.9j ['$' for help]"})
		if !serial_success {
			fmt.Println("Failed to setup gantry serial communication... retrying")
		}
	}

	fmt.Println("Successfully setup gantry serial communication")

	gantry.disableLock()

	return gantry
}

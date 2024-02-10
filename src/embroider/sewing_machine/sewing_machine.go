package sewing_machine

import (
	"embroider/serial"
	"fmt"
	"time"
)

type SewingMachine struct {
	port serial.Serial
}

type State int64

const (
	UNKNOWN State = 0
	IDLE    State = 1
	RUNNING State = 2
)

func (s *SewingMachine) getStatus() (State, bool) {
	success, incoming := s.port.HandleRequest([]string{"?"}, 1, 3)
	if !success {
		return UNKNOWN, false
	}

	if incoming[0] == "IDLE" {
		return IDLE, true
	}

	return RUNNING, true
}

func (s *SewingMachine) Pulse(retry int) bool {
	for {
		if !s.port.HandleCommand([]serial.Transmission{
			serial.Transmission{
				Out: []string{"PULSE"},
				In:  []string{"ACK"},
			},
		}, 3) {
			fmt.Printf("Sewing Machine: Failed to send pulse command\n")
			continue
		}

		status, success := s.getStatus()

		for status == RUNNING && success {
			time.Sleep(250 * time.Millisecond)
			status, success = s.getStatus()
		}

		if !success {
			fmt.Printf("Sewing Machine: Failed to get status\n")
			continue
		}

		fmt.Printf("Successfully completed sewing machine operation\n")
		return true
	}

	return false
}

func Create(port string) SewingMachine {
	var sewing_machine SewingMachine

	serial_success := false
	for !serial_success {
		sewing_machine.port, serial_success = serial.SetupSerialCommunication(port, []string{"?"}, []string{"IDLE"})
		if !serial_success {
			fmt.Println("Failed to setup sewing machine serial communication... retrying")
		}
	}

	fmt.Println("Successfully setup sewing machine serial communication")

	return sewing_machine
}

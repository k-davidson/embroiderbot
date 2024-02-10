package serial

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

type CommsErrorCode int

var SUCCESS CommsErrorCode = 0
var TIMEOUT CommsErrorCode = 1
var FATAL CommsErrorCode = 2

type CommsChannel struct {
	port           serial.Port
	shutdownSignal chan bool
	message        chan string
	errorCode      chan CommsErrorCode
}

type Serial struct {
	write CommsChannel
	read  CommsChannel
}

func (s *Serial) Close() {
	s.write.shutdownSignal <- true
	s.read.shutdownSignal <- true
	s.write.port.Close()
}

func (s *Serial) HandleRequest(requests []string, expected_responses int, timeout time.Duration) (bool, []string) {
	for _, out := range requests {
		s.write.message <- out
		select {
		case err := <-s.write.errorCode:
			switch err {
			case SUCCESS:
				continue
			case TIMEOUT:
				fmt.Printf("Communication failure: Write timeout on \"%s\"\n", out)
				return false, []string{}
			case FATAL:
				fmt.Printf("Communication failure: Write fatal on \"%s\"\n", out)
				return false, []string{}
			}
		}
	}

	var result []string

	for i := 0; i < expected_responses; i++ {
		select {
		case incoming := <-s.read.message:
			result = append(result, incoming)
		case <-time.After(timeout * time.Second):
			fmt.Printf("Communication failure: Timeout\n")
			return false, []string{}
		}
	}

	return true, result
}

func (s *Serial) HandleCommand(sequence []Transmission, timeout time.Duration) bool {
	for _, transmission := range sequence {
		success, incoming := s.HandleRequest(transmission.Out, len(transmission.In), timeout)
		if !success {
			return false
		}

		if len(incoming) != len(transmission.In) {
			fmt.Printf("Communication failure: Number of responses does not match expected\n")
			return false
		}

		for i, in := range incoming {
			if in != transmission.In[i] {
				fmt.Printf("Communication failure: Response failure, expected \"%s\", got \"%s\"\n", transmission.In[i], in)
				return false
			}
		}
	}

	return true
}

func readThread(channel CommsChannel) {
	buff := make([]byte, 100)
	type ReadResult struct {
		err   error
		bytes int
	}
	read_chan := make(chan ReadResult, 1)
	var command string

	for {
		go func() {
			var result ReadResult
			result.bytes, result.err = channel.port.Read(buff)
			read_chan <- result
		}()
		select {
		case <-channel.shutdownSignal:
			return
		case read_result := <-read_chan:
			if read_result.err != nil {
				log.Fatal(read_result.err)
			}
			if read_result.bytes == 0 {
				fmt.Println("\nEOF")
			}

			command += strings.TrimSuffix(string(buff[:read_result.bytes]), "\r")
			if strings.HasPrefix(command, "\r\n") {
				command = command[2:]
			}
			for strings.Contains(command, "\n") {
				index := strings.Index(command, "\n")
				channel.message <- command[:index-1]
				if index+1 == len(command) {
					command = ""
				} else {
					command = command[index+1:]
				}
			}
		}
	}
}

func writeThread(channel CommsChannel) {
	type WriteResult struct {
		err   error
		bytes int
	}
	write_chan := make(chan WriteResult, 1)
	for {
		select {
		case <-channel.shutdownSignal:
			return
		case msg := <-channel.message:
			go func() {
				var result WriteResult
				result.bytes, result.err = channel.port.Write([]byte(msg + "\n\r"))
				write_chan <- result
			}()
			select {
			case write_result := <-write_chan:
				if write_result.err != nil {
					log.Fatal(write_result.err)
				} else {
					channel.errorCode <- SUCCESS
				}
			case <-time.After(3 * time.Second):
				channel.errorCode <- TIMEOUT
			}
		}
	}
}

func GetSerialPortSelection() string {
	valid_selection := false

	var selection int

	ports, err := serial.GetPortsList()

	for !valid_selection {
		if err != nil {
			log.Fatal(err)
		}
		if len(ports) == 0 {
			log.Fatal("No serial ports found!")
		}
		fmt.Println("Please select a port to connect to by index:")
		for i, port := range ports {
			fmt.Printf("[%d] Found port: %v\n", i, port)
		}

		fmt.Print("> ")
		_, err = fmt.Scanln(&selection)
		if err != nil {
			fmt.Println("Error taking input.\n")
			continue
		} else if len(ports) < selection {
			fmt.Println("Port selection out of range.")
		} else {
			valid_selection = true
		}
	}

	return ports[selection]
}

func SetupSerialCommunication(port_name string, start_outgoing []string, start_incoming []string) (Serial, bool) {
	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(port_name, mode)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(3.0 * time.Second)

	read_channel := CommsChannel{
		port:           port,
		shutdownSignal: make(chan bool, 1),
		message:        make(chan string, 1),
		errorCode:      make(chan CommsErrorCode, 1),
	}

	go readThread(read_channel)

	write_channel := CommsChannel{
		port:           port,
		shutdownSignal: make(chan bool, 1),
		message:        make(chan string, 1),
		errorCode:      make(chan CommsErrorCode, 1),
	}

	go writeThread(write_channel)

	s := Serial{write_channel, read_channel}

	expected_comms := []Transmission{Transmission{Out: start_outgoing, In: start_incoming}}
	if !s.HandleCommand(expected_comms, 3) {
		s.Close()
		return Serial{}, false
	}

	return s, true
}

type Transmission struct {
	Out []string
	In  []string
}

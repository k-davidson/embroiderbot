package main

import (
	"embroider/gantry"
	"embroider/gcode"
	"embroider/serial"
	"embroider/sewing_machine"
	"fmt"
	"time"
	"os"
	"bufio"
)

func getInput(input chan string) {
    for {
        in := bufio.NewReader(os.Stdin)
        result, err := in.ReadString('\n')
        if err != nil {
            continue
        }

        input <- result
    }
}

func main() {
	var sequences [3]gcode.GCodeSequence
	sequences[0] = gcode.ParseGcodeFile("/Users/kelbiedavidson/Desktop/bitmaps/test/koala0.gcode")
	sequences[1] = gcode.ParseGcodeFile("/Users/kelbiedavidson/Desktop/bitmaps/test/koala1.gcode")
	sequences[2] = gcode.ParseGcodeFile("/Users/kelbiedavidson/Desktop/bitmaps/test/koala2.gcode")




	for i, _ := range(sequences) {
		sequences[i].Scale(0.1)
	}

	trans := sequences[0].Min()
	for i := 1; i < len(sequences); i++ {
		next_min := sequences[i].Min()
		for j, _ := range(trans) {
			if next_min[j] < trans[j] {
				trans[j] = next_min[j]
			}
		}
	}

	for i, _ := range(trans) {
		trans[i] = -trans[i]
	}

	for i, _ := range(sequences) {
		sequences[i].Translate(trans)
		sequence_max := sequences[i].Max()
		sequence_min := sequences[i].Min()
		fmt.Printf("%.2f %.2f to %.2f %.2f\n", sequence_min[0], sequence_min[1], sequence_max[0], sequence_max[1])
	}

	for _, sequence := range(sequences) {
		for _, code := range sequence.Sequence() {
			fmt.Printf("%s\n", code.ToString())
		}
	}
	
	gantry := gantry.Create(serial.GetSerialPortSelection())
	sewing_machine := sewing_machine.Create(serial.GetSerialPortSelection())

	retry := 10

	input := make(chan string, 1)
	go getInput(input)

	iterations := 0
	average_time := float32(0)

	for j, sequence := range(sequences) {
		fmt.Printf("Commencing sequence %d.\n", j);
		for i, code := range sequence.Sequence() {
			start := time.Now()
			fmt.Printf("%s (%d of %d)\n", code.ToString(), i, len(sequence.Sequence()))
			if !gantry.HandleGcode(code, retry) {
				fmt.Printf("Failed to position gantry %d times. Exiting.", retry)
				return
			}
			select {
				case in := <-input:
					if (in == "P\n") {
						fmt.Printf("Paused...\n")
						for {
							in := <- input;
							if (in == "P\n") {
								fmt.Printf("Starting..\n.")
								break;
							}
						}
					}
				case <-time.After(10 * time.Millisecond):
					;
			}

			if !sewing_machine.Pulse(retry) {
				fmt.Printf("Failed to pulse sewing machine %d times. Exiting.", retry)
				return
			}

			if i != 0 {
				elapsed := time.Since(start)
				average_time += float32(elapsed.Milliseconds()) / 1000
				iterations += 1

				fmt.Printf("%.2f seconds to complete\n", average_time / float32(iterations))
			}
		}

		fmt.Printf("Complete sequence %d. Press any key to continue...\n", j);
		_ = <-input
	}

	return

	// scanner := bufio.NewScanner(os.Stdin)
	// fmt.Printf("> ")
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	if line == "PULSE" {
	// 		sewing_machine.Pulse()
	// 	} else if code := gcode.GCodefactory(line); code != nil {
	// 		gantry.HandleGcode(code)
	// 	} else {
	// 		fmt.Printf("ERROR: GCode \"%s\" not handled\n", line)
	// 	}
	// 	fmt.Printf("> ")
	// }
}

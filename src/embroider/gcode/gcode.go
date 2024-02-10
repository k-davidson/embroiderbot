package gcode

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type GCode interface {
	FromString([]string)
	ToString() string
	Position() [3]float32
	Translate([3]float32)
	scale(float32)
	same(GCode) bool
}

type G01 struct {
	code     int
	position [3]float32
	rate     int
}

func (code G01) same(comp GCode) bool {
	v, ok := comp.(*G01)
	if !ok {
		return false;
	}

	for i, _ := range(code.position) {
		if code.position[i] != v.position[i] {
			return false;
		}
	}

	if code.rate != v.rate {
		return false;
	}

	return true
}

func (code G01) ToString() string {
	return fmt.Sprintf("G%d X%.2f Y%.2f Z%.2f F%d", code.code, code.position[0], code.position[1], code.position[2], code.rate)
}

func (code G01) Position() [3]float32 {
	return code.position
}

func (code *G01) Translate(delta [3]float32) {
	code.position[0] = code.position[0] + delta[0]
	code.position[1] = code.position[1] + delta[1]
	code.position[2] = code.position[2] + delta[2]
}

func (code *G01) scale(scale float32) {
	code.position[0] = code.position[0] * scale
	code.position[1] = code.position[1] * scale
	code.position[2] = code.position[2] * scale
}

func (code *G01) FromString(in []string) {
	code.code = 1
	fmt.Sscanf(in[0], "X%g", &code.position[0])
	fmt.Sscanf(in[1], "Y%g", &code.position[1])
	fmt.Sscanf(in[1], "Z%g", &code.position[2])
	fmt.Sscanf(in[3], "F%d", &code.rate)
}

type GCodeSequence struct {
	sequence []GCode
	min_pos  [2]float32
	max_pos  [2]float32
	idx      int
}

func (seq *GCodeSequence) append(code GCode) {
	pos := code.Position()
	seq.min_pos[0] = min(pos[0], seq.min_pos[0])
	seq.min_pos[1] = min(pos[1], seq.min_pos[1])
	seq.max_pos[0] = max(pos[0], seq.max_pos[0])
	seq.max_pos[1] = max(pos[1], seq.max_pos[1])

	seq.sequence = append(seq.sequence, code)
}

func (seq GCodeSequence) Min() [3]float32 {
	min_bounds := seq.sequence[0].Position()
	for _, code := range seq.sequence {
		for i, val := range code.Position() {
			if val < min_bounds[i] {
				min_bounds[i] = val
			}
		}
	}

	return min_bounds
}

func (seq GCodeSequence) Max() [3]float32 {
	max_bounds := seq.sequence[0].Position()
	for _, code := range seq.sequence {
		for i, val := range code.Position() {
			if max_bounds[i] < val {
				max_bounds[i] = val
			}
		}
	}

	return max_bounds
}

func (seq *GCodeSequence) Translate(delta [3]float32) {
	for _, code := range seq.sequence {
		code.Translate(delta)
	}
}

func (seq *GCodeSequence) Zero() {
	min_bounds := seq.Min()
	dx := -min_bounds[0]
	dy := -min_bounds[1]
	fmt.Printf("%.2f %.2f dx/dy\n", dx, dy)

	seq.Translate([3]float32{dx, dy, 0})

	max_bounds := seq.Max()
	min_bounds = seq.Min()
	fmt.Printf("%.2f %.2f and %.2f %.2f\n", min_bounds[0], min_bounds[1], max_bounds[0], max_bounds[1])
}

func (seq *GCodeSequence) Scale(scalar float32) {
	for _, code := range seq.sequence {
		code.scale(scalar)
	}
}

func (seq GCodeSequence) Contains(code GCode) bool {
	for _, existing := range(seq.sequence) {
		if (existing.same(code)) {
			return true;
		}
	}

	return false;
}

func (seq GCodeSequence) Sequence() []GCode {
	return seq.sequence
}

func GCodefactory(in string) GCode {
	components := strings.Split(in, " ")

	switch components[0] {
	case "G1":
		var code G01
		code.FromString(components[1:])
		return &code
	}

	return nil
}

func ParseGcodeFile(filepath string) GCodeSequence {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}

	var sequence GCodeSequence

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if gcode := GCodefactory(scanner.Text()); gcode != nil {
			sequence.append(gcode)
		}

	}

	return sequence
}

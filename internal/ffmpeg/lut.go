package ffmpeg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CubeLUT is a 3D lookup table parsed from an Adobe/Iridas .cube file.
type CubeLUT struct {
	Title     string
	Size      int // grid size per axis (2..256); the table holds Size^3 entries
	DomainMin [3]float64
	DomainMax [3]float64
	Table     []float32 // Size^3 RGB triplets in [0,1]; red channel varies fastest
}

// ParseCubeFile reads and validates a 3D .cube LUT from disk.
func ParseCubeFile(path string) (*CubeLUT, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read LUT file: %v", err)
	}
	return ParseCube(data)
}

// ParseCube parses the raw contents of a 3D .cube LUT file. Errors are
// descriptive and include the offending line number.
func ParseCube(data []byte) (*CubeLUT, error) {
	lut := &CubeLUT{
		DomainMin: [3]float64{0, 0, 0},
		DomainMax: [3]float64{1, 1, 1},
	}

	size := 0
	var values []float32

	content := strings.TrimPrefix(string(data), "\ufeff") // tolerate UTF-8 BOM

	for i, rawLine := range strings.Split(content, "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(strings.TrimRight(rawLine, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		keyword := strings.ToUpper(fields[0])

		switch keyword {
		case "TITLE":
			title := strings.TrimSpace(strings.TrimPrefix(line, fields[0]))
			lut.Title = strings.Trim(title, "\"")

		case "LUT_3D_SIZE":
			if len(fields) != 2 {
				return nil, fmt.Errorf("line %d: LUT_3D_SIZE expects exactly one value", lineNo)
			}
			n, err := strconv.Atoi(fields[1])
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid LUT_3D_SIZE %q", lineNo, fields[1])
			}
			if n < 2 || n > 256 {
				return nil, fmt.Errorf("line %d: LUT_3D_SIZE %d out of range (2..256)", lineNo, n)
			}
			size = n

		case "LUT_1D_SIZE", "LUT_1D_INPUT_RANGE":
			return nil, fmt.Errorf("line %d: only 3D LUTs are supported", lineNo)

		case "LUT_3D_INPUT_RANGE":
			if len(fields) != 7 {
				return nil, fmt.Errorf("line %d: LUT_3D_INPUT_RANGE expects 6 values", lineNo)
			}
			for j := 0; j < 3; j++ {
				minV, err := strconv.ParseFloat(fields[1+j], 64)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid LUT_3D_INPUT_RANGE value %q", lineNo, fields[1+j])
				}
				maxV, err := strconv.ParseFloat(fields[4+j], 64)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid LUT_3D_INPUT_RANGE value %q", lineNo, fields[4+j])
				}
				lut.DomainMin[j] = minV
				lut.DomainMax[j] = maxV
			}

		case "DOMAIN_MIN", "DOMAIN_MAX":
			if len(fields) != 4 {
				return nil, fmt.Errorf("line %d: %s expects 3 values", lineNo, keyword)
			}
			var v [3]float64
			for j := 0; j < 3; j++ {
				f, err := strconv.ParseFloat(fields[j+1], 64)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid %s value %q", lineNo, keyword, fields[j+1])
				}
				v[j] = f
			}
			if keyword == "DOMAIN_MIN" {
				lut.DomainMin = v
			} else {
				lut.DomainMax = v
			}

		default:
			// Anything else must be table data: whitespace-separated RGB triplets.
			if size == 0 {
				return nil, fmt.Errorf("line %d: table data found before LUT_3D_SIZE", lineNo)
			}
			if len(fields)%3 != 0 {
				return nil, fmt.Errorf("line %d: expected RGB triplets, found %d values", lineNo, len(fields))
			}
			for _, tok := range fields {
				f, err := strconv.ParseFloat(tok, 32)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid LUT value %q", lineNo, tok)
				}
				if f < 0 || f > 1 {
					return nil, fmt.Errorf("line %d: LUT value %v out of range (0..1)", lineNo, f)
				}
				values = append(values, float32(f))
			}
		}
	}

	if size == 0 {
		return nil, fmt.Errorf("missing LUT_3D_SIZE header")
	}
	for j := 0; j < 3; j++ {
		if lut.DomainMax[j] <= lut.DomainMin[j] {
			return nil, fmt.Errorf("DOMAIN_MAX must be greater than DOMAIN_MIN")
		}
	}

	expected := 3 * size * size * size
	if len(values) != expected {
		return nil, fmt.Errorf("expected %d LUT entries (%d^3 RGB triplets), found %d", expected/3, size, len(values)/3)
	}

	lut.Size = size
	lut.Table = values
	return lut, nil
}

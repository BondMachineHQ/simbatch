package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Config struct {
	workingDir       string
	inputFile        string
	outputFile       string
	simulationSteps  string
	isML             bool
	benchCore        bool
	stopOnValidOf    int
	linearDataRange  string
	dataType         string
	prefix           string
	header           bool
	omitPrefix       bool
	delaysFile       string
	delayString      string
}

func usage() {
	fmt.Println("SimBatch: A batch simulator for BondMachine designs")
	fmt.Println("")
	fmt.Println("SimBatch allows you to run batch simulations of BondMachine designs using a CSV input file and producing a CSV output file.")
	fmt.Println("It expects the BondMachine design to be already compiled in the working directory called `bondmachine.json`.")
	fmt.Println("")
	fmt.Println("Usage: simbatch [options]")
	fmt.Println("Options:")
	fmt.Println("  -w, --working-dir DIR        set the working directory (default: working_dir)")
	fmt.Println("  -i, --input-file FILE        set the input CSV file (default: simbatch_input.csv)")
	fmt.Println("  -o, --output-file FILE       set the output CSV file (default: working_dir/simbatch_output.csv)")
	fmt.Println("  -s, --simulation-steps N     number of simulation steps (default: 200)")
	fmt.Println("  -m, --ml                     enable ML output formatting (probabilities + classification)")
	fmt.Println("  -b, --benchcore              enable benchcore mode")
	fmt.Println("  -H, --header                 include header row in output CSV")
	fmt.Println("  -P, --prefix                 include data type prefix in output CSV")
	fmt.Println("  -d, --data-type TYPE         data type for outputs (e.g. float32) (default: float32)")
	fmt.Println("  -l, --linear-data-range RANGE    pass a linear data range option to bondmachine/bmnumbers")
	fmt.Println("  -v, --stop-on-valid-of N     stop on valid of output index N")
	fmt.Println("  -h, --help                   show this help message and exit")
	fmt.Println("  -y, --delays-file FILE       set the delays file")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("  simbatch -w working_dir -i input.csv -o out.csv -s 200")
}

func runCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

func loadInputsOrOutputs(workingDir, option, linearDataRange string) (map[string]string, error) {
	command := fmt.Sprintf("bondmachine -bondmachine-file %s/bondmachine.json %s %s", workingDir, option, linearDataRange)
	output, err := runCommand(command)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result, scanner.Err()
}

func getPrefix(dataType, linearDataRange string) (string, error) {
	command := fmt.Sprintf("bmnumbers -get-prefix %s %s", dataType, linearDataRange)
	output, err := runCommand(command)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func findMaxIndex(values []float32) int {
	if len(values) == 0 {
		return -1
	}
	maxIdx := 0
	maxVal := values[0]
	for i, v := range values {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

func main() {
	cfg := Config{
		simulationSteps: "200",
		dataType:        "float32",
		prefix:          "0f",
		omitPrefix:      true,
		stopOnValidOf:   -1,
	}

	// Parse command-line flags
	flag.StringVar(&cfg.workingDir, "w", "", "working directory")
	flag.StringVar(&cfg.workingDir, "working-dir", "", "working directory")
	flag.StringVar(&cfg.inputFile, "i", "", "input CSV file")
	flag.StringVar(&cfg.inputFile, "input-file", "", "input CSV file")
	flag.StringVar(&cfg.outputFile, "o", "", "output CSV file")
	flag.StringVar(&cfg.outputFile, "output-file", "", "output CSV file")
	flag.StringVar(&cfg.simulationSteps, "s", "200", "simulation steps")
	flag.StringVar(&cfg.simulationSteps, "simulation-steps", "200", "simulation steps")
	flag.BoolVar(&cfg.isML, "m", false, "enable ML output")
	flag.BoolVar(&cfg.isML, "ml", false, "enable ML output")
	flag.BoolVar(&cfg.benchCore, "b", false, "enable benchcore")
	flag.BoolVar(&cfg.benchCore, "benchcore", false, "enable benchcore")
	flag.BoolVar(&cfg.header, "H", false, "include header")
	flag.BoolVar(&cfg.header, "header", false, "include header")
	var includePrefix bool
	flag.BoolVar(&includePrefix, "P", false, "include prefix")
	flag.BoolVar(&includePrefix, "prefix", false, "include prefix")
	flag.StringVar(&cfg.dataType, "d", "float32", "data type")
	flag.StringVar(&cfg.dataType, "data-type", "float32", "data type")
	var linearDataRangeArg string
	flag.StringVar(&linearDataRangeArg, "l", "", "linear data range")
	flag.StringVar(&linearDataRangeArg, "linear-data-range", "", "linear data range")
	flag.IntVar(&cfg.stopOnValidOf, "v", -1, "stop on valid of")
	flag.IntVar(&cfg.stopOnValidOf, "stop-on-valid-of", -1, "stop on valid of")
	flag.StringVar(&cfg.delaysFile, "y", "", "delays file")
	flag.StringVar(&cfg.delaysFile, "delays-file", "", "delays file")
	helpFlag := flag.Bool("h", false, "show help")
	helpFlagLong := flag.Bool("help", false, "show help")

	flag.Parse()

	if *helpFlag || *helpFlagLong {
		usage()
		os.Exit(0)
	}

	// Set defaults
	if cfg.workingDir == "" {
		cfg.workingDir = "working_dir"
	}
	if cfg.inputFile == "" {
		cfg.inputFile = "simbatch_input.csv"
	}
	if cfg.outputFile == "" {
		cfg.outputFile = cfg.workingDir + "/simbatch_output.csv"
	}
	if linearDataRangeArg != "" {
		cfg.linearDataRange = "-linear-data-range " + linearDataRangeArg
	}
	if cfg.delaysFile != "" {
		cfg.delayString = "-sim-delays-file " + cfg.delaysFile
	}
	if includePrefix {
		cfg.omitPrefix = false
	}

	// Load BondMachine inputs
	inputs, err := loadInputsOrOutputs(cfg.workingDir, "-list-inputs", cfg.linearDataRange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading inputs: %v\n", err)
		os.Exit(1)
	}

	// Load BondMachine outputs
	outputs, err := loadInputsOrOutputs(cfg.workingDir, "-list-outputs", cfg.linearDataRange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading outputs: %v\n", err)
		os.Exit(1)
	}

	// Get number prefix
	prefix, err := getPrefix(cfg.dataType, cfg.linearDataRange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting prefix: %v\n", err)
		os.Exit(1)
	}
	cfg.prefix = prefix

	// Open input file
	inputFileHandle, err := os.Open(cfg.inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer inputFileHandle.Close()

	// Create output file
	outputFileHandle, err := os.Create(cfg.outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outputFileHandle.Close()

	writer := bufio.NewWriter(outputFileHandle)
	defer writer.Flush()

	// Write header if needed
	if cfg.header {
		if cfg.isML {
			for i := 0; i < len(outputs); i++ {
				fmt.Fprintf(writer, "probability_%d,", i)
			}
			fmt.Fprint(writer, "classification")
		}
		if cfg.benchCore {
			fmt.Fprint(writer, ",latency_cycles")
		}
		fmt.Fprintln(writer)
	}

	// Process each line
	scanner := bufio.NewScanner(inputFileHandle)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		inputsValues := strings.Split(line, ",")
		if len(inputsValues) == 0 {
			continue
		}

		if len(inputsValues) != len(inputs) {
			fmt.Fprintf(os.Stderr, "Error: The input file has an invalid number of columns\n")
			continue
		}

		fmt.Printf("Running simulation with inputs: %s\n", line)

		// Remove simbox file
		simboxFile := cfg.workingDir + "/simboxtemp.json"
		os.Remove(simboxFile)

		// Prepare simbox general commands
		commands := []string{
			fmt.Sprintf("simbox -simbox-file %s -add \"config:show_io_pre\"", simboxFile),
			fmt.Sprintf("simbox -simbox-file %s -suspend 0", simboxFile),
			fmt.Sprintf("simbox -simbox-file %s -add \"config:show_io_post\"", simboxFile),
			fmt.Sprintf("simbox -simbox-file %s -suspend 1", simboxFile),
			fmt.Sprintf("simbox -simbox-file %s -add \"config:show_ticks\"", simboxFile),
			fmt.Sprintf("simbox -simbox-file %s -suspend 2", simboxFile),
			fmt.Sprintf("simbox -simbox-file %s -add \"config:show_pc\"", simboxFile),
			fmt.Sprintf("simbox -simbox-file %s -suspend 3", simboxFile),
			fmt.Sprintf("simbox -simbox-file %s -add \"config:show_disasm\"", simboxFile),
			fmt.Sprintf("simbox -simbox-file %s -suspend 4", simboxFile),
		}

		for _, cmd := range commands {
			if _, err := runCommand(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error preparing simbox command: %s\n%v\n", cmd, err)
				os.Exit(2)
			}
		}

		// Add inputs to simbox
		for i := 0; i < len(inputsValues); i++ {
			inputName := inputs[strconv.Itoa(i)]
			inputValue := inputsValues[i]
			cmd := fmt.Sprintf("simbox -simbox-file %s -add \"absolute:0:set:%s:%s\"", simboxFile, inputName, inputValue)
			if _, err := runCommand(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error setting input %s to %s\n%v\n", inputName, inputValue, err)
				os.Exit(2)
			}
		}

		// Prepare outputs
		for outputName, outputVal := range outputs {
			var cmd string
			if cfg.benchCore && outputVal == "o"+strconv.Itoa(len(outputs)-1) {
				cmd = fmt.Sprintf("simbox -simbox-file %s -add \"onexit:show:%s:unsigned\"", simboxFile, outputVal)
			} else {
				cmd = fmt.Sprintf("simbox -simbox-file %s -add \"onexit:show:%s:%s\"", simboxFile, outputVal, cfg.dataType)
			}
			if _, err := runCommand(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error getting output %s\n%v\n", outputName, err)
				os.Exit(2)
			}
		}

		cfg.stopOnValidOf = len(outputs) - 1

		// Run simulation
		simCmd := fmt.Sprintf("bondmachine -bondmachine-file %s/bondmachine.json %s -simbox-file %s -sim-stop-on-valid-of %d -sim -sim-interactions %s %s",
			cfg.workingDir, cfg.delayString, simboxFile, cfg.stopOnValidOf, cfg.simulationSteps, cfg.linearDataRange)
		simOutput, err := runCommand(simCmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running simulation\n%v\n", err)
			os.Exit(2)
		}

		outline := strings.TrimSpace(simOutput)
		if cfg.omitPrefix {
			outline = strings.ReplaceAll(outline, cfg.prefix, "")
		}

		// Handle benchCore mode
		var latencyCycles string
		if cfg.benchCore {
			parts := strings.Fields(outline)
			if len(parts) > 0 {
				latencyCycles = parts[len(parts)-1]
				outline = strings.Join(parts[:len(parts)-1], " ")
			}
		}

		// Handle ML mode
		if cfg.isML {
			vals := strings.Fields(outline)
			floatVals := make([]float32, len(vals))
			for i, v := range vals {
				f, _ := strconv.ParseFloat(v, 32)
				floatVals[i] = float32(f)
			}
			maxIdx := findMaxIndex(floatVals)
			outline = strings.ReplaceAll(outline, " ", ",")
			outline = fmt.Sprintf("%s,%d", outline, maxIdx)
		} else {
			outline = strings.Trim(outline, ",")
			outline = strings.ReplaceAll(outline, " ", ",")
		}

		// Write output
		fmt.Fprint(writer, outline)
		if cfg.benchCore {
			fmt.Fprintf(writer, ",%s", latencyCycles)
		}
		fmt.Fprintln(writer)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}
}

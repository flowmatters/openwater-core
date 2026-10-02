package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	_ "path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/util/logger"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

const templatePath = "pre/ow-specgen/*.got"
const floatTemplate = "[+-]?([0-9]*[.])?[0-9]+"
const parameterTemplate = `(\[(?P<Min>%s),(?P<Max>%s)\](?P<Units>[a-zA-Z/%%^0-9·.\-]*)?)?\s*(?P<Description>[^,]*)(,\s*default=(?P<Default>%s))?`
const dimensionTemplate = `\[([_a-zA-Z][_a-zA-Z0-9]*)(,([_a-zA-Z][_a-zA-Z0-9]*))*\]`

type ModelSpecs map[string]ModelSpec
type VariableSpec struct {
	Name           string
	Units          string
	Position       int
	Default        float64
	Description    string
	Range          []float64
	IsDimension    bool
	Dimensions     []string
	Dimensionality int
}

type ModelSpec struct {
	Filename       string
	Name           string
	Symbol         string
	Reference      string
	Package        string
	Inputs         yaml.MapSlice
	InputSpecs     []VariableSpec
	States         yaml.MapSlice
	Dimensions     []string
	Parameters     yaml.MapSlice
	ParameterSpecs []VariableSpec
	Outputs        yaml.MapSlice
	OutputSpecs    []VariableSpec
	Implementation yaml.MapSlice
	Init           yaml.MapSlice
	ExtractStates  yaml.MapSlice
	Reports        yaml.MapSlice
	ReportSpecs    []ReportSpec
	Flags          struct {
		GenerateStruct        bool
		GenerateVector        bool
		GenerateInit          bool
		GenerateExtractStates bool
		ZeroStates            bool
		PassOutputsAsParams   bool
		GenerateReports       bool
	}
	SingleFunc        string
	InitFunc          string
	PackStatsFunc     string
	ExtractStatesFunc string
}

// ReportSpec is one entry from a model's reports: block, ready for codegen.
type ReportSpec struct {
	Name         string
	ConstName    string
	Code         int
	Severity     string // the sim.Sev* constant to emit
	SeverityName string
	Template     string
	NArgs        int
}

var (
	// '[warning] release rate not converged -- est {0}, avg {1} m^3.s^-1'
	reportSeverityRe    = regexp.MustCompile(`^\s*\[(?P<Severity>[a-zA-Z]+)\]\s*(?P<Template>.*)$`)
	reportPlaceholderRe = regexp.MustCompile(`\{(\d+)\}`)
	reportNameRe        = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
)

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func severityConstant(s sim.Severity) (string, error) {
	switch s {
	case sim.SevWarning:
		return "sim.SevWarning", nil
	case sim.SevError:
		return "sim.SevError", nil
	}
	return "", fmt.Errorf("severity %s cannot be declared on a report", s)
}

// parseReports turns a model's reports: block into codegen-ready specs.
//
// Report messages are free-form English, which makes them the most
// parse-hazardous thing in an OW-SPEC block: an unquoted value containing ": "
// is YAML structure, not text. Rather than let that through as a confusing
// type error, it is called out by name here.
func parseReports(desc ModelSpec) ([]ReportSpec, error) {
	if len(desc.Reports) == 0 {
		return nil, nil
	}

	if len(desc.Implementation) == 0 {
		return nil, fmt.Errorf("model %s declares reports but has no implementation block", desc.Name)
	}

	seen := make(map[string]struct{}, len(desc.Reports))
	result := make([]ReportSpec, 0, len(desc.Reports))

	for i, entry := range desc.Reports {
		name := fmt.Sprint(entry.Key)

		if !reportNameRe.MatchString(name) {
			return nil, fmt.Errorf("model %s: report name %q must be a Go-style identifier", desc.Name, name)
		}
		if _, dup := seen[name]; dup {
			return nil, fmt.Errorf("model %s: duplicate report %q", desc.Name, name)
		}
		seen[name] = struct{}{}

		text, ok := entry.Value.(string)
		if !ok {
			return nil, fmt.Errorf("model %s: report %q must be a quoted string -- "+
				"an unquoted message containing \": \" or starting with \"{\" is read as YAML "+
				"structure rather than text (parsed as %T)", desc.Name, name, entry.Value)
		}

		matches := reportSeverityRe.FindStringSubmatch(text)
		if matches == nil {
			return nil, fmt.Errorf("model %s: report %q must begin with a bracketed severity, "+
				"e.g. '[warning] message with {0}'", desc.Name, name)
		}

		severity, err := sim.ParseSeverity(matches[1])
		if err != nil {
			return nil, fmt.Errorf("model %s: report %q: %w", desc.Name, name, err)
		}
		severityConst, err := severityConstant(severity)
		if err != nil {
			return nil, fmt.Errorf("model %s: report %q: %w", desc.Name, name, err)
		}

		template := strings.TrimSpace(matches[2])
		if template == "" {
			return nil, fmt.Errorf("model %s: report %q has a severity but no message", desc.Name, name)
		}

		nArgs := 0
		for _, placeholder := range reportPlaceholderRe.FindAllStringSubmatch(template, -1) {
			idx, err := strconv.Atoi(placeholder[1])
			if err != nil {
				return nil, fmt.Errorf("model %s: report %q: bad placeholder %s", desc.Name, name, placeholder[0])
			}
			if idx >= sim.MaxReportArgs {
				return nil, fmt.Errorf("model %s: report %q uses {%d}, but a report carries at most "+
					"%d arguments ({0} to {%d})", desc.Name, name, idx, sim.MaxReportArgs, sim.MaxReportArgs-1)
			}
			if idx+1 > nArgs {
				nArgs = idx + 1
			}
		}

		result = append(result, ReportSpec{
			Name:         name,
			ConstName:    fmt.Sprintf("%sReport%s", desc.Name, capitalise(name)),
			Code:         i,
			Severity:     severityConst,
			SeverityName: severity.String(),
			Template:     template,
			NArgs:        nArgs,
		})
	}

	return result, nil
}

func processPath(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		return processDirectory(path)
	case mode.IsRegular():
		return processFile(path)
	}

	return nil
}

func processDirectory(path string) error {
	return nil
}

func transform(spec ModelSpec) ModelSpec {
	spec.ParameterSpecs = make([]VariableSpec, len(spec.Parameters))
	spec.Dimensions = make([]string, 0)
	dims := make(map[string]struct{})

	for i, v := range spec.Parameters {
		paramName := fmt.Sprint(v.Key)
		var dimensions []string = nil
		if strings.Contains(paramName, "[") {
			cleanName := strings.Replace(strings.Replace(paramName, "[", ",", 1), "]", "", 1)
			nameComponents := strings.Split(cleanName, ",")
			paramName = nameComponents[0]
			dimensions = nameComponents[1:]
			for _, d := range dimensions {
				dims[d] = struct{}{}
			}
			log.Info().Any(paramName, dimensions).Msg("")
		}

		txt := fmt.Sprint(v.Value)
		if txt == `<nil>` {
			txt = ""
		}

		_ = regexp.MustCompile(floatTemplate)
		r := regexp.MustCompile(fmt.Sprintf(parameterTemplate, floatTemplate, floatTemplate, floatTemplate))
		matches := r.FindStringSubmatch(txt)

		units := matches[6]
		description := matches[7]
		defaultVal, err := strconv.ParseFloat(matches[9], 64)
		if err != nil {
			defaultVal = 0.0
		}
		minVal, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			minVal = 0.0
		}
		maxVal, err := strconv.ParseFloat(matches[4], 64)
		if err != nil {
			maxVal = 0.0
		}
		paramRange := make([]float64, 2)
		paramRange[0] = minVal
		paramRange[1] = maxVal

		spec.ParameterSpecs[i] = VariableSpec{
			paramName,
			units,
			i,
			defaultVal,
			description,
			paramRange,
			false,
			dimensions,
			len(dimensions) + 1}
	}

	for _, v := range spec.ParameterSpecs {
		if v.Dimensions == nil {
			continue
		}
		for _, d := range v.Dimensions {
			for i, p := range spec.ParameterSpecs {
				if strings.Compare(p.Name, d) == 0 {
					log.Info().Str("Dimension", p.Name).Str("Variable", v.Name).Msg("Variable dimension")
					spec.ParameterSpecs[i].IsDimension = true
				}
			}
		}
	}

	for k := range dims {
		spec.Dimensions = append(spec.Dimensions, k)
	}

	for _, v := range spec.ParameterSpecs {
		if v.IsDimension {
			log.Info().Str("Dimension", v.Name).Msg("Is a dimension")
		}
	}

	// Parse input specs
	spec.InputSpecs = make([]VariableSpec, len(spec.Inputs))
	for i, v := range spec.Inputs {
		name := fmt.Sprint(v.Key)
		units := ""
		if v.Value != nil {
			units = fmt.Sprint(v.Value)
		}
		spec.InputSpecs[i] = VariableSpec{
			Name:  name,
			Units: units,
		}
	}

	// Parse output specs
	spec.OutputSpecs = make([]VariableSpec, len(spec.Outputs))
	for i, v := range spec.Outputs {
		name := fmt.Sprint(v.Key)
		units := ""
		if v.Value != nil {
			units = fmt.Sprint(v.Value)
		}
		spec.OutputSpecs[i] = VariableSpec{
			Name:  name,
			Units: units,
		}
	}

	return spec
}

func processFile(fn string) error {
	//	directory := filepath.Dir(fn)

	contents, err := os.ReadFile(fn)
	if err != nil {
		return fmt.Errorf("reading %s: %w", fn, err)
	}

	// (?m) so a file with a licence header or package doc comment above the
	// package clause is still processed. Without it the clause had to be at
	// byte 0, and any other file was skipped silently -- generating nothing.
	packageRe := regexp.MustCompile("(?m)^package\\s+(\\w+)")
	packageMatch := packageRe.FindSubmatch(contents)
	if packageMatch == nil {
		log.Info().Str("Filename", fn).Msg("No package declaration found")
		return nil
	}
	packageName := packageMatch[1]

	re := regexp.MustCompile("(?smU)/(\\*\\s*OW-SPEC)(.*)(\\*/)")

	specs := re.FindAllSubmatch(contents, -1)
	var failures error
	for _, specContents := range specs {
		if specContents == nil {
			continue
		}

		tabRe := regexp.MustCompile("\t")
		spec := specContents[2]
		spec = tabRe.ReplaceAll(spec, []byte("  "))

		desc := make(ModelSpecs) // := make(map[interface{}]interface{})
		err = yaml.Unmarshal(spec, &desc)
		if err != nil {
			log.Error().Stack().Err(err).Str("Filename", fn).Msg("Invalid OW-SPEC block")
			failures = errors.Join(failures, fmt.Errorf("parsing OW-SPEC in %s: %w", fn, err))
			continue
		}

		for name, description := range desc {
			if description.Package == "" {
				description.Package = string(packageName)
			}

			if description.Name == "" {
				description.Name = name
			}

			description = transform(description)

			description.Filename = fn
			if err := generateWrapper(description); err != nil {
				log.Error().Stack().Err(err).Str("Filename", fn).Str("Model", description.Name).Msg("Could not generate wrapper")
				failures = errors.Join(failures, err)
			}
		}
	}

	return failures
}

func generateWrapper(desc ModelSpec) error {
	reports, err := parseReports(desc)
	if err != nil {
		return err
	}
	desc.ReportSpecs = reports
	desc.Flags.GenerateReports = len(reports) > 0

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"inc": func(n int) int {
			return n + 1
		},
		"lower": strings.ToLower}).ParseGlob(templatePath)

	if err != nil {
		log.Fatal().Stack().Err(err).Msg("Could not parse templates. Exiting")
	}
	tmpl = tmpl.Funcs(template.FuncMap{
		"inc": func(n int) int {
			return n + 1
		}})

	dir := filepath.Dir(desc.Filename)
	destFn := filepath.Join(dir, fmt.Sprintf("generated_%s.go", desc.Name))
	log.Info().Str("Destination Filename", destFn).Msg("Writing generated wrapper")
	dest, err := os.Create(destFn)
	if err != nil {
		return fmt.Errorf("creating %s: %w", destFn, err)
	}
	defer dest.Close()

	if len(desc.Implementation) > 0 {
		desc.Flags.GenerateStruct = true

		for _, e := range desc.Implementation {
			if e.Key == "type" && e.Value == "scalar" {
				desc.Flags.GenerateVector = true
			}

			if e.Key == "function" {
				desc.SingleFunc = e.Value.(string)
			}

			if e.Key == "outputs" {
				desc.Flags.PassOutputsAsParams = e.Value.(string) == "params"
			}
		}
	}

	for _, e := range desc.Init {
		if e.Key == "type" && e.Value == "scalar" {
			desc.Flags.GenerateInit = true
		}

		if e.Key == "zero" && e.Value == true {
			desc.Flags.ZeroStates = true
		}

		if e.Key == "function" {
			desc.InitFunc = e.Value.(string)
		}
	}

	desc.Flags.GenerateExtractStates = true
	for _, e := range desc.ExtractStates {
		if e.Key == "function" {
			desc.Flags.GenerateExtractStates = false
			desc.ExtractStatesFunc = e.Value.(string)
		}
		if e.Key == "packfunc" {
			desc.PackStatsFunc = e.Value.(string)
		}
	}

	err = tmpl.ExecuteTemplate(dest, "generated_struct.got", desc)
	if err != nil {
		// The wrapper on disk is now truncated. Remove it so a later build
		// can't compile against a half-written file.
		dest.Close()
		os.Remove(destFn)
		return fmt.Errorf("generating %s: %w", destFn, err)
	}

	return nil
}

func main() {
	flag.Parse()
	logger.SetupLogger(false, true)
	paths := flag.Args()

	exitCode := 0
	for _, path := range paths {
		if err := processPath(path); err != nil {
			log.Error().Stack().Err(err).Str("Path", path).Msg("Failed to process path")
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}

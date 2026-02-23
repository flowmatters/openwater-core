package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	_ "path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/flowmatters/openwater-core/util/logger"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

const templatePath = "pre/ow-specgen/*.got"
const floatTemplate = "[+-]?([0-9]*[.])?[0-9]+"
const parameterTemplate = `(\[(?P<Min>%s),(?P<Max>%s)\]((?P<Units>[\s]+))?)?\s*(?P<Description>[^,]*)(,\s*default=(?P<Default>%s))?`
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
	Package        string
	Inputs         yaml.MapSlice
	States         yaml.MapSlice
	Dimensions     []string
	Parameters     yaml.MapSlice
	ParameterSpecs []VariableSpec
	Outputs        yaml.MapSlice
	Implementation yaml.MapSlice
	Init           yaml.MapSlice
	ExtractStates  yaml.MapSlice
	Flags          struct {
		GenerateStruct        bool
		GenerateVector        bool
		GenerateInit          bool
		GenerateExtractStates bool
		ZeroStates            bool
		PassOutputsAsParams   bool
	}
	SingleFunc        string
	InitFunc          string
	PackStatsFunc     string
	ExtractStatesFunc string
}

func processPath(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		log.Error().Stack().Err(err).Msg("")
		return
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		processDirectory(path)
	case mode.IsRegular():
		processFile(path)
	}
}

func processDirectory(path string) {

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

		units := matches[7]
		description := matches[8]
		defaultVal, err := strconv.ParseFloat(matches[10], 64)
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
	return spec
}

func processFile(fn string) {
	//	directory := filepath.Dir(fn)

	contents, err := os.ReadFile(fn)
	if err != nil {
		log.Error().Stack().Err(err).Msg("")
		return
	}

	packageRe := regexp.MustCompile("^\\s*package\\s+(\\w+)")
	packageMatch := packageRe.FindSubmatch(contents)
	if packageMatch == nil {
		log.Info().Str("Filename", fn).Msg("No package declaration found")
		return
	}
	packageName := packageMatch[1]

	re := regexp.MustCompile("(?smU)/(\\*\\s*OW-SPEC)(.*)(\\*/)")

	specs := re.FindAllSubmatch(contents, -1)
	for _, specContents := range specs {
		if specContents == nil {
			return
		}

		tabRe := regexp.MustCompile("\t")
		spec := specContents[2]
		spec = tabRe.ReplaceAll(spec, []byte("  "))

		desc := make(ModelSpecs) // := make(map[interface{}]interface{})
		err = yaml.Unmarshal(spec, &desc)
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
			return
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
			generateWrapper(description)
		}
	}
}

func generateWrapper(desc ModelSpec) {
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
		log.Error().Stack().Err(err).Msg("")
		return
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
		log.Error().Stack().Err(err).Msg("")
		return
	}
}

func main() {
	flag.Parse()
	logger.SetupLogger(false, true)
	paths := flag.Args()

	for _, path := range paths {
		processPath(path)
	}
}

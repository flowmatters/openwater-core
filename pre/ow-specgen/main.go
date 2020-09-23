package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	_ "path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

const templatePath = "pre/ow-specgen/*.got"
const floatTemplate = "[+-]?([0-9]*[.])?[0-9]+"
const parameterTemplate = `(\[(?P<Min>%s),(?P<Max>%s)\]((?P<Units>[\s]+))?)?\s*(?P<Description>[^,]*)(,\s*default=(?P<Default>%s))?`
const dimensionTemplate = `\[([_a-zA-Z][_a-zA-Z0-9]*)(,([_a-zA-Z][_a-zA-Z0-9]*))*\]`

type ModelSpecs map[string]ModelSpec
type VariableSpec struct {
	Name        string
	Units       string
	Position    int
	Default     float64
	Description string
	Range       []float64
	IsDimension bool
	Dimensions  []string
	Dimensionality int
}

type ModelSpec struct {
	Filename       string
	Name           string
	Package        string
	Inputs         yaml.MapSlice
	States         yaml.MapSlice
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
		fmt.Println(err)
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
	for i, v := range spec.Parameters {
		paramName := fmt.Sprint(v.Key)
		var dimensions []string = nil
		if strings.Contains(paramName,"[") {
			cleanName := strings.Replace(strings.Replace(paramName,"[",",",1),"]","",1)
			nameComponents := strings.Split(cleanName,",")
			paramName = nameComponents[0]
			dimensions = nameComponents[1:]
			fmt.Println(paramName,dimensions)
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

	for _,v := range spec.ParameterSpecs {
		if v.Dimensions == nil {
			continue
		}
		for _,d := range v.Dimensions {
			for i,p := range spec.ParameterSpecs {
				if strings.Compare(p.Name,d) == 0 {
					fmt.Printf("%s is a dimension of %s\n",p.Name,v.Name)
				  spec.ParameterSpecs[i].IsDimension = true
				}
			}
		}
	}

	for _,v := range spec.ParameterSpecs {
		if v.IsDimension {
			fmt.Printf("%s is a dimension\n",v.Name)
		}
	}
	return spec
}

func processFile(fn string) {
	//	directory := filepath.Dir(fn)

	contents, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Println(err)
		return
	}

	packageRe := regexp.MustCompile("^\\s*package\\s+(\\w+)")
	packageMatch := packageRe.FindSubmatch(contents)
	if packageMatch == nil {
		fmt.Printf("No package declaration in %s\n", fn)
		return
	}
	packageName := packageMatch[1]

	re := regexp.MustCompile("(?smU)/(\\*\\s*OW-SPEC)(.*)(\\*/)")

	specContents := re.FindSubmatch(contents)
	if specContents == nil {
		//		fmt.Printf("No OW-SPEC in %s\n",fn)
		return
	}

	tabRe := regexp.MustCompile("\t")
	spec := specContents[2]
	spec = tabRe.ReplaceAll(spec, []byte("  "))

	desc := make(ModelSpecs) // := make(map[interface{}]interface{})
	err = yaml.Unmarshal(spec, &desc)
	if err != nil {
		fmt.Println(err)
		// for i := range specContents {
		// 	fmt.Println(i)
		// 	fmt.Println(string(specContents[i]))
		// }
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
		// fmt.Println(name)
		// fmt.Println(description)
		generateWrapper(description)
	}
}

func generateWrapper(desc ModelSpec) {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"inc": func(n int) int {
			return n + 1
		},
		"lower": strings.ToLower}).ParseGlob(templatePath)

	if err != nil {
		fmt.Println(err)
		fmt.Println("Could not parse templates. Exiting")
		os.Exit(1)
	}
	//	fmt.Println(tmpl.DefinedTemplates())
	tmpl = tmpl.Funcs(template.FuncMap{
		"inc": func(n int) int {
			return n + 1
		}})

	dir := filepath.Dir(desc.Filename)
	destFn := filepath.Join(dir, fmt.Sprintf("generated_%s.go", desc.Name))
	fmt.Printf("Writing to %s\n", destFn)
	dest, err := os.Create(destFn)
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
		return
	}
	// var implementationField = desc.Implementation
	// var generateStruct = implementationField != nil

	// if generateStruct {
	// 	implementation, _ := implementationField.(map[interface{}]interface{})

	// 		if implementation==nil {
	// 			fmt.Println("invalid type")
	// 			return
	// 		}

	// 		function, _ := implementation["function"].(string)
	// 		funcType, _ := implementation["type"].(string)
	// 		lang, _ := implementation["lang"].(string)
	// 		fmt.Println(function,funcType,lang)
	// } else {
	// 	fmt.Println("no implementation specified. assuming type already exists")
	// }

	// Extract out yaml (convert tabs to spaces if necessary)
	// Unmarshal object
	// If implementation is C function - generate Go wrapper
	// If implementation is single function - generate vectorised wrapper
	// If implementation is function - generate type
	// Generate description

	// Write out
}

func main() {
	flag.Parse()
	paths := flag.Args()

	for _, path := range paths {
		processPath(path)
	}
}

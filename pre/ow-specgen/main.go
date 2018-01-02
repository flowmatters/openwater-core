package main

import (
	"strings"
	"path/filepath"
	"regexp"
	"io/ioutil"
	"fmt"
  "gopkg.in/yaml.v2"
	"flag"
	"os"
	_	"path/filepath"
	"text/template"
)

const templatePath = "src/github.com/flowmatters/openwater-core/pre/ow-specgen/*.got"

type ModelSpecs map[string]ModelSpec
type VariableSpec struct {
	Name string
	Units string
	Position int
	Default float64
	Description string
}

type ModelSpec struct {
	Filename string
	Name string
	Package string
	Inputs yaml.MapSlice
	States yaml.MapSlice
	Parameters yaml.MapSlice
	ParameterSpecs []VariableSpec
	Outputs yaml.MapSlice
	Implementation yaml.MapSlice
	Init yaml.MapSlice
	ExtractStates yaml.MapSlice
	Flags struct {
		GenerateStruct bool
		GenerateVector bool
		GenerateInit bool
		GenerateExtractStates bool
		ZeroStates bool
		PassOutputsAsParams bool
	}
	SingleFunc string
	InitFunc string
	PackStatsFunc string
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

func processDirectory(path string){

}

func transform(spec ModelSpec) ModelSpec {
	spec.ParameterSpecs = make([]VariableSpec,len(spec.Parameters))
	for i,v := range spec.Parameters {
		spec.ParameterSpecs[i] = VariableSpec{v.Key.(string),"",i,0,""}
	}
	return spec
}

func processFile(fn string) {
//	directory := filepath.Dir(fn)

	contents,err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Println(err)
		return
	}

	packageRe := regexp.MustCompile("^\\s*package\\s+(\\w+)")
	packageMatch := packageRe.FindSubmatch(contents)
	if packageMatch == nil {
		fmt.Printf("No package declaration in %s\n",fn)
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
	spec = tabRe.ReplaceAll(spec,[]byte("  "))

	desc := make(ModelSpecs) // := make(map[interface{}]interface{})
	err = yaml.Unmarshal(spec,&desc)
	if err != nil {
		fmt.Println(err)
		// for i := range specContents {
		// 	fmt.Println(i)
		// 	fmt.Println(string(specContents[i]))
		// }
		return
	}

	for name,description := range desc{
		if description.Package=="" {
			description.Package = string(packageName)
		}

		if description.Name=="" {
			description.Name = name
		}

		description = transform(description)

		description.Filename = fn
		// fmt.Println(name)
		// fmt.Println(description)
		generateWrapper(description)
	}
}

func generateWrapper(desc ModelSpec){
	tmpl,err := template.New("").Funcs(template.FuncMap{
		"inc": func(n int) int {
				return n + 1
		},
		"lower":strings.ToLower}).ParseGlob(templatePath)

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
	destFn := filepath.Join(dir,fmt.Sprintf("generated_%s.go",desc.Name))
	fmt.Printf("Writing to %s\n",destFn)
	dest, err := os.Create(destFn)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dest.Close()

	if len(desc.Implementation)>0 {
		desc.Flags.GenerateStruct = true

		for _,e := range(desc.Implementation){
			if e.Key=="type" && e.Value=="scalar" {
				desc.Flags.GenerateVector = true
			}

			if e.Key=="function"{
				desc.SingleFunc = e.Value.(string)
			}

			if e.Key=="outputs"{
				desc.Flags.PassOutputsAsParams = e.Value.(string)=="params"
			}
		}
	}

	for _,e := range(desc.Init) {
		if e.Key=="type" && e.Value=="scalar" {
			desc.Flags.GenerateInit = true
		}

		if e.Key=="zero" && e.Value==true {
			desc.Flags.ZeroStates = true
		}

		if e.Key=="function"{
			desc.InitFunc = e.Value.(string)
		}
	}

	desc.Flags.GenerateExtractStates = true
	for _,e := range(desc.ExtractStates) {
		if e.Key=="function"{
			desc.Flags.GenerateExtractStates = false
			desc.ExtractStatesFunc = e.Value.(string)
		}
		if e.Key=="packfunc" {
			desc.PackStatsFunc = e.Value.(string)
		}
	}

	err = tmpl.ExecuteTemplate(dest,"generated_struct.got",desc)
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

func main(){
	flag.Parse()
	paths := flag.Args()

	for _,path := range(paths) {
		processPath(path)
	}
}
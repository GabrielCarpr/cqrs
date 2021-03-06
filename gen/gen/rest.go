package gen

import (
	"bytes"
	"errors"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/GabrielCarpr/cqrs/gen/templates"
	"gopkg.in/yaml.v2"
)

type group struct {
	Path       string   `yaml:"path"`
	Middleware []string `yaml:"middleware,flow"`
	Groups     []group  `yaml:"groups,flow"`
	Routes     []route  `yaml:"routes,flow"`
}

func (g group) valid() error {
	for _, route := range g.Routes {
		if err := route.valid(); err != nil {
			return err
		}
	}
	for _, group := range g.Groups {
		if err := group.valid(); err != nil {
			return err
		}
	}
	return nil
}

func (g group) imports() []string {
	names := []string{}
	for _, route := range g.Routes {
		switch true {
		case route.Command != "":
			names = append(names, route.Command)
		case route.Query.Answer != "":
			names = append(names, route.Query.Answer)
			names = append(names, route.Query.Question)
		case route.Event != "":
			names = append(names, route.Event)
		}
	}
	for _, mw := range g.Middleware {
		names = append(names, mw)
	}
	for _, grp := range g.Groups {
		names = append(names, grp.imports()...)
	}
	return names
}

func (g group) Imports() map[string]string {

	return imports(g.imports()...)
}

type route struct {
	Path    string `yaml:"path"`
	Method  string `yaml:"method"`
	Command string `yaml:"command"`
	Query   struct {
		Question string `yaml:"question"`
		Answer   string `yaml:"answer"`
		Adapter  string `yaml:"adapter"`
	} `yaml:"query"`
	Event      string   `yaml:"event"`
	Async      bool     `yaml:"async"`
	Middleware []string `yaml:"middleware,flow"`
}

func (r route) valid() error {
	selected := 0
	switch true {
	case r.Command != "":
		selected++
	case r.Query.Question != "" && r.Query.Answer != "":
		selected++
	case r.Event != "":
		selected++
	}

	switch {
	case r.Path == "":
		return errors.New("Path must be specified")
	case r.Method == "":
		return errors.New("Method must be specified")
	case selected != 1:
		return errors.New("Command, query, or event must be provided for route")
	case r.Query.Answer != "" && r.Query.Question == "":
		return errors.New("Question cannot be empty when answer provided")
	case r.Query.Question != "" && r.Query.Answer == "":
		return errors.New("Answer cannot be empty when question provided")
	case strings.Contains(r.Query.Adapter, "."):
		return errors.New("Adapter must be in the same package")
	}
	return nil
}

func server(name string) string {
	pkg := packageName(name)
	if pkg == homePkg {
		return "server"
	}
	return alias(name)
}

func rest(routesPath string) {
	filename := strings.Replace(filepath.Base(routesPath), ".yml", "", 1)
	r, err := ioutil.ReadFile(routesPath)
	if err != nil {
		log.Fatal(err)
	}

	config := group{}
	err = yaml.Unmarshal(r, &config)
	if err != nil {
		log.Fatal(err)
	}
	if err := config.valid(); err != nil {
		log.Fatal(err)
	}

	templ, err := template.New("rest.go.tmpl").Funcs(map[string]interface{}{
		"packageName": packageName,
		"structName":  structName,
		"alias":       alias,
		"server":      server,
	}).ParseFS(templates.Templates, "rest.go.tmpl", "restGroup.go.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewBuffer([]byte{})
	err = templ.Execute(buf, config)
	if err != nil {
		log.Fatal(err)
	}
	output := filepath.Join(".", filename+"_gen.go")

	err = ioutil.WriteFile(output, buf.Bytes(), fs.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

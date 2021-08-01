package make

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/GabrielCarpr/cqrs/gen/templates"
)

func init() {
	addTemplate("command", "command.go.tmpl")
	addTemplate("query", "query.go.tmpl")
	addTemplate("test", "test.go.tmpl")
}

var targets map[string]*template.Template = make(map[string]*template.Template)

func addTemplate(name string, file string) {
	tmpl, err := template.New(file).ParseFS(templates.Templates, file)
	if err != nil {
		panic(err)
	}
	targets[name] = tmpl
}

type name struct {
	Name string
}

func (n name) Valid() error {
	switch {
	case !strings.Contains(n.Name, "."):
		return errors.New("Name must be in form `path/to/package.Name")
	}
	return nil
}

func (n name) Struct() string {
	parts := strings.Split(n.Name, ".")
	return parts[len(parts)-1]
}

func (n name) Path() string {
	abs := strings.Split(n.Name, ".")[0]
	return strings.Join(strings.Split(abs, "/")[1:], "/")
}

func (n name) Package() string {
	path := n.Path()
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func (n name) Filename(t string) string {
	return fmt.Sprintf("%s_%s.go", strings.ToLower(n.Struct()), t)
}

func Make(args ...string) {
	target := strings.ToLower(args[0])
	targetName := name{args[1]}
	if err := targetName.Valid(); err != nil {
		log.Fatal(err)
	}

	tmpl, ok := targets[target]
	if !ok {
		log.Fatalf("Cannot create `%s`", target)
	}

	if err := os.MkdirAll(targetName.Path(), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(filepath.Join(targetName.Path(), targetName.Filename(target)))
	if err != nil {
		log.Fatal(err)
	}

	if err := tmpl.Execute(f, targetName); err != nil {
		log.Fatal(err)
	}
}

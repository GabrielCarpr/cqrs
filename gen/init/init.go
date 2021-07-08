package init

import (
	"bytes"
	"errors"
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/GabrielCarpr/cqrs/gen/templates"
)

var (
	makeCmd = flag.NewFlagSet("make", flag.ExitOnError)
)

type context struct {
	Name string
	Root string
}

type tmplMap map[string]*template.Template

type fileMap map[string]*bytes.Buffer

func (c context) Validate() error {
	switch true {
	case c.Name == "":
		return errors.New("Name must be provided")
	}
	return nil
}

func Init(args ...string) {
	c := context{}
	makeCmd.StringVar(&c.Name, "name", "", "The name of the application to create")
	makeCmd.StringVar(&c.Root, "root", ".", "The root directory to initialize into")

	err := makeCmd.Parse(args)
	if err != nil {
		log.Fatal(err)
	}
	err = c.Validate()
	if err != nil {
		log.Fatal(err)
	}

	res := render(c.Root, getTemplates(), c)
	save(c.Root, res)
}

func getTemplates() tmplMap {
	tm := make(tmplMap)
	fsys := templates.Templates
	fs.WalkDir(fsys, "init", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		tName := strings.Replace(path, "init/", "", -1)

		data, err := fsys.ReadFile(path)
		if err != nil {
			return err
		}
		tm[tName] = template.Must(template.New(tName).Parse(string(data)))
		return nil
	})
	return tm
}

func render(root string, tm tmplMap, c context) fileMap {
	res := make(fileMap)
	for path, tmpl := range tm {
		res[path] = &bytes.Buffer{}
		err := tmpl.Execute(res[path], c)
		if err != nil {
			log.Fatal(err)
		}
	}
	return res
}

func save(root string, fm fileMap) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		log.Fatal(err)
	}

	for path, data := range fm {
		target := strings.Replace(filepath.Join(absRoot, path), ".tmpl", "", 1)
		err := os.MkdirAll(filepath.Dir(target), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		f, err := os.Create(target)
		if err != nil {
			log.Fatal(err)
		}
		f.Write(data.Bytes())
	}
}

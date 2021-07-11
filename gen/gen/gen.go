package gen

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

type gqlRoutes struct {
	Commands []gqlCommand `yaml:",flow"`
	Queries  []gqlQuery   `yaml:",flow"`
}

func (r gqlRoutes) Imports() map[string]string {
	imports := make(map[string]string)
	for _, c := range r.Commands {
		imports[c.Package()] = c.Alias()
	}
	for _, q := range r.Queries {
		imports[q.Package("query")] = q.Alias("query")
		imports[q.Package("output")] = q.Alias("output")
	}
	return imports
}

type gqlCommand struct {
	Command string
	Sync    bool
}

func (c gqlCommand) Package() string {
	spl := strings.Split(c.Command, ".")
	return spl[0]
}

func (c gqlCommand) Struct() string {
	spl := strings.Split(c.Command, ".")
	return spl[1]
}

func (c gqlCommand) Alias() string {
	pkg := c.Package()
	hash := sha256.Sum256([]byte(pkg))
	str := hex.EncodeToString(hash[:])
	r := regexp.MustCompile(`[0-9]`)
	return r.ReplaceAllString(str, "")[:8]
}

type gqlQuery struct {
	Query   string
	Output  string
	Adapter *string
	Return  *string
}

func (q gqlQuery) Slice() bool {
	return strings.Contains(q.Output, "[]")
}

func (q gqlQuery) Package(item string) string {
	var spl []string
	switch item {
	case "query":
		spl = strings.Split(q.Query, ".")
		break
	case "output":
		spl = strings.Split(q.Output, ".")
	}

	return spl[0]
}

func (q gqlQuery) Struct(item string) string {
	var spl []string
	switch item {
	case "query":
		spl = strings.Split(q.Query, ".")
		break
	case "output":
		spl = strings.Split(q.Output, ".")
		break
	}

	return strings.Replace(spl[1], "[]", "", -1)
}

func (q gqlQuery) Alias(item string) string {
	pkg := q.Package(item)
	hash := sha256.Sum256([]byte(pkg))
	str := hex.EncodeToString(hash[:])
	r := regexp.MustCompile(`[0-9]`)
	return r.ReplaceAllString(str, "")[:8]
}

func (q gqlQuery) CodeName(item string) string {
	n := fmt.Sprintf("%s.%s", q.Alias(item), q.Struct(item))
	if q.Slice() {
		return "[]" + n
	}
	return n
}

func Graphql() {
	routes, err := ioutil.ReadFile("../graphql/routes.yml")
	if err != nil {
		log.Fatal(err)
	}
	result := gqlRoutes{}

	err = yaml.Unmarshal(routes, &result)
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.ParseFiles("./templates/adapters.go.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buf, result)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("../graphql/adapters.go", buf.Bytes(), 1)
	if err != nil {
		log.Fatal(err)
	}
}

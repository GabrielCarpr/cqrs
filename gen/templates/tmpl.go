package templates

import "embed"

//go:embed */*.sh *.go.tmpl init/go.mod.tmpl init/*/*.go.tmpl init/*/*/*.go.tmpl init/*.go.tmpl init/migrations/*.sql
var Templates embed.FS

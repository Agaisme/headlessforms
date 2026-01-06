package web

import (
	"embed"
)

//go:embed build/*
var StaticFiles embed.FS

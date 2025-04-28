package internal

import "embed"

//go:embed python-libs/data/*
var PythonLibs embed.FS

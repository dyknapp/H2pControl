package internal

import "embed"

//go:embed assets/python-libs/data/**
var PythonLibs embed.FS

//go:embed assets/protoc/**
var ProtocBinaries embed.FS

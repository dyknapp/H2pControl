package internal

import "embed"

//go:embed assets/protoc/**
var ProtocBinaries embed.FS

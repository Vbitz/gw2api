package web

import "embed"

// Assets contains all embedded web assets (templates, CSS, JS)
//go:embed assets/templates/* assets/templates/partials/* assets/static/css/* assets/static/js/* assets/static/*
var Assets embed.FS
package embed

import "embed"

//go:embed api/openapi.json
//go:embed api/openapi.yml
var SwaggerFS embed.FS

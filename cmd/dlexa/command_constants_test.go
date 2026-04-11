package main

import "testing"

const (
	unexpectedErrorFormat         = "unexpected error: %v"
	executeRootCommandErrorFormat = "executeRootCommand() error = %v"
	moduleWantDPDFormat           = "module = %q, want dpd"
)

func TestCommandConstants(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		got  string
		want string
	}{
		"dpd command":              {got: commandDPD, want: "dpd"},
		"search command":           {got: commandSearch, want: "search"},
		"noticia command":          {got: commandNoticia, want: "noticia"},
		"duda-linguistica command": {got: commandDudaLinguistica, want: "duda-linguistica"},
		"espanol-al-dia command":   {got: commandEspanolAlDia, want: "espanol-al-dia"},
		"duda-linguistica help":    {got: helpCommandDudaLinguistica, want: "dlexa duda-linguistica"},
		"espanol-al-dia help":      {got: helpCommandEspanolAlDia, want: "dlexa espanol-al-dia"},
		"dpd syntax":               {got: syntaxDPD, want: "dlexa dpd <termino>"},
		"dpd search syntax":        {got: syntaxDPDSearch, want: "dlexa dpd search <termino-de-busqueda>"},
		"noticia syntax":           {got: syntaxNoticia, want: "dlexa noticia <slug>"},
		"duda-linguistica syntax":  {got: syntaxDudaLinguistica, want: "dlexa duda-linguistica <slug>"},
		"espanol-al-dia syntax":    {got: syntaxEspanolAlDia, want: "dlexa espanol-al-dia <slug>"},
	}

	for name, tc := range tests {
		if tc.got != tc.want {
			t.Fatalf("%s = %q, want %q", name, tc.got, tc.want)
		}
	}
}

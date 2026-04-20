package app

const (
	errExecuteModule         = "ExecuteModule() error = %v"
	errLookupType            = "lookup type = %T, want *query.LookupService"
	errRegistryType          = "registry type = %T, want *source.StaticRegistry"
	errSourcesFor            = "SourcesFor() error = %v"
	errResolvedSourcesLen    = "resolved sources len = %d, want 1"
	errPipelineSourceType    = "source type = %T, want *source.PipelineSource"
	commandDudaLinguistica   = "duda-linguistica"
	commandEspanolAlDia      = "espanol-al-dia"
	sourceBusquedaGeneralRAE = "búsqueda general RAE"
	helpSyntaxDlexaRoot      = "dlexa <comando> [argumentos]"
)

package engine

import (
	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

// ParseInput is the shared parser-engine execution envelope.
type ParseInput struct {
	Descriptor model.SourceDescriptor
	Document   fetch.Document
}

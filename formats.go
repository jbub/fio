package fio

// ExportFormat represents export formats supported by Export.
type ExportFormat string

// Supported ExportFormat types.
const (
	JSONFormat ExportFormat = "json"
	XMLFormat  ExportFormat = "xml"
	CSVFormat  ExportFormat = "csv"
	GPCFormat  ExportFormat = "gpc"
	HTMLFormat ExportFormat = "html"
	OFXFormat  ExportFormat = "ofx"
)

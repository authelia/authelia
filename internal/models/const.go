package models

const (
	errFmtValueNil           = "cannot value model type '%T' with value nil to driver.Value"
	errFmtScanNil            = "cannot scan model type '%T' from value nil: type doesn't support nil values"
	errFmtScanInvalidType    = "cannot scan model type '%T' from type '%T' with value '%v'"
	errFmtScanInvalidTypeErr = "cannot scan model type '%T' from type '%T' with value '%v': %w"
)

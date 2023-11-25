package pumps

import "github.com/nico612/iam-demo/internal/pump/analytics"

// CommonPumpConfig defines common options used by
type CommonPumpConfig struct {
	filters               analytics.AnalyticsFilters
	timeout               int
	OmitDetailedRecording bool
}

// SetFilters  set attributes `filters` for CommonPumpConfig.
func (p *CommonPumpConfig) SetFilters(filters analytics.AnalyticsFilters) {
	p.filters = filters
}

// GetFilters get attributes `filters` for CommonPumpConfig.
func (p *CommonPumpConfig) GetFilters() analytics.AnalyticsFilters {
	return p.filters
}

// SetTimeout set attributes `timeout` for CommonPumpConfig.
func (p *CommonPumpConfig) SetTimeout(timeout int) {
	p.timeout = timeout
}

// GetTimeout get attributes `timeout` for CommonPumpConfig.
func (p *CommonPumpConfig) GetTimeout() int {
	return p.timeout
}

// SetOmitDetailedRecording set attributes `OmitDetailedRecording` for CommonPumpConfig.
func (p *CommonPumpConfig) SetOmitDetailedRecording(omitDetailRecording bool) {
	p.OmitDetailedRecording = omitDetailRecording
}

// GetOmitDetailedRecording get attributes `OmitDetailedRecording` for CommonPumpConfig.
func (p *CommonPumpConfig) GetOmitDetailedRecording() bool {
	return p.OmitDetailedRecording
}

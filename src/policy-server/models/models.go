package models

type Policy struct {
	Source      Source      `json:"source"`
	Destination Destination `json:"destination"`
}

type Source struct {
	ID  string `json:"id"`
	Tag string `json:"tag,omitempty"`
}

type Destination struct {
	ID       string `json:"id"`
	Tag      string `json:"tag,omitempty"`
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

type Tag struct {
	ID  string `json:"id"`
	Tag string `json:"tag"`
}

type PoliciesFilter struct {
	SourceGuids      []string `json:source_guids`
	DestinationGuids []string `json:destination_guids`
}

type Space struct {
	Name    string `json:name`
	OrgGUID string `json:organization_guid`
}

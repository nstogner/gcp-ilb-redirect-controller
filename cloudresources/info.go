package cloudresources

// Info is the information needed to CRUD resources.
type Info struct {
	Project string
	Region  string

	GeneratedName string

	IP       string
	Hostname string
}

package service

const (
	UUIDv4Regex                    = `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	AlphanumericAndUnderscoreRegex = `^[a-zA-Z0-9_]+$`
	HostnameRegexStringRFC1123     = `^([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62}){1}(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?$`
)

var (
	OltpDatabaseTypes = []string{
		"Oracle",
		"MSSQL",
		"MySQL",
		"Postgres",
	}
)

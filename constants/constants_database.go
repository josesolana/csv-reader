package constants

const (
	// DbDriver Database Driver
	DbDriver = "postgres"
	//DbUser Database User
	DbUser = "postgres"
	//DbName Database Name
	DbName = "customers_dev"
	//DbNameTest Database Name
	DbNameTest = "customers_test"
	//DbPass Database Password
	DbPass = "postgres"

	// TotalRetry Number of time before skip a row
	TotalRetry = 3

	//IDPos Position into DB
	IDPos = 0
	//IsProcessedPos Position into DB
	IsProcessedPos = 1
	//RetryPos Position into DB
	RetryPos = 2
)

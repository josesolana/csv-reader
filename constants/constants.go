package constants

import "time"

const (
	// DbDriver Database Driver
	DbDriver = "postgres"
	//DbUser Database User
	DbUser = "postgres"
	//DbName Database Name
	DbName = "postgres"
	//DbPass Database Password
	DbPass = "postgres"
	//Columns Default columns order
	Columns = "id, first_name, last_name, email, phone"
	// AcceptedExt File extension accepted
	AcceptedExt = ".csv"
	//Workers Number of go routines concurrently
	Workers = 30
	// Buff Workers's buffer channel
	Buff = 15
	// BatchSizeRow Batch to no overload
	BatchSizeRow = 200
	//CRMUrl CRM Json Appi's Url. This is a example server
	CRMUrl = "https://jsonplaceholder.typicode.com/posts"
	//CRMUrlFAIL CRM Json Appi's Url. This is a example server FAILING
	CRMUrlFail = "https://jsonplaceholder.typicode.com/posts/FAIL"
	// TotalRetry Retry Number before skip the row
	TotalRetry = 20
	//TimeOut to Http requests
	TimeOut = time.Duration(3 * time.Second)
)

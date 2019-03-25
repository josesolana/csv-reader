package constants

import "time"

const (
	// AcceptedExt File extension accepted
	AcceptedExt = ".csv"

	//Workers Number of go routines concurrently
	Workers = 30
	// Buff Workers's buffer channel
	Buff = 15
	// BatchSizeRow Batch to no overload and also prevent block it all, to run more than consumer(integrator)
	BatchSizeRow = 200

	//CRMUrl CRM Json Appi's Url. This is a example server
	CRMUrl = "https://jsonplaceholder.typicode.com/posts"
	//CRMUrlFAIL CRM Json Appi's Url. This is a example server FAILING
	CRMUrlFail = "https://jsonplaceholder.typicode.com/posts/FAIL"
	//TimeOut to Http requests
	TimeOut = time.Duration(3 * time.Second)
)

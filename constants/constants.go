package constants

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
	Workers = 20
	// Buff Workers's buffer channel
	Buff = 5
)

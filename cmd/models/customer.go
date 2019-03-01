package models

// CustomerTableName Customer Table Name into DB
var CustomerTableName = "customers"

// CustomerColumns Columns in order to make query
var CustomerColumns = "id, first_name, last_name, email, phone"

// Customer Customer's database model.
type Customer struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

# csv-reader
CSV reader and CRM integrator.

## Summary

Let’s assume we have a large csv file that contains the information of the customers we work with. The CSV can be very large and the information will not fit in the system memory. Once we have the data parsed we need to save it in the database (Postgresql) and send this information to a CRM (Customer
Relation Management) service.

The CRM has a standard JSON API, but it is not very reliable and may fail randomly. We need to make sure that we send the data to the CRM as quickly as possible while ensuring we don’t send the same customer data to the CRM system multiple times.

As we expect the application to grow, we split the project into two small services the CSV reader and CRM integrator. We need to create two go binaries, both the programs will run on the same machine. You can decide the best mode of communication between the programs.

Please share an architectural diagram of the system along with the code for both the services.

Use the go standard library to build the project.
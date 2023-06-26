# financial-app
A simple microservice which handles financial transactions. The transaction service exposes a RESTful API (Application Programming Interface), allowing users to make financial transactions between two bank accounts and enforcing various acceptance criteria.

# Code Structure
## pkg
The "pkg" folder will have all the code that could be used in other packages. In our case, we have a subfolder called "models", which will have the domain definitions for the account and transaction data models. That is all the struct and interface definitions, which will describe the data entities.
## internal
All the private code, which cannot be imported into other applications, will go into the internal package. Each application will have a separate subfolder within it. There's a subfolder for the `transaction` application. In our case, it has multiple subfolders, which include:
- rest: All code implements the REST-based APIs
- repo: A simple key-value store for the accounts and transactions.
- service: Any specific business logic required for the application. In our case, we locate the transaction service.
## mocks
To be done

# Usage
1. Clone the repository or download the code files.
2. Open the terminal and navigate to the project directory.
3. Install the necessary dependencies by running `go mod download`.
4. Run `go build` to compile the packages and their dependencies.
5. Run the service through `go run financial-app`.
   By default, the microservice runs on http://localhost:8000. You can test the API using tools like cURL or Postman.

# Testing
To run the tests for the microservice, execute the following command:
```
go test -v
```

# API Endpoints
## Perform a transaction
**Endpoint:** `/transactions`

**Method:** POST

**Request Body:**
```
{
  "source_account_id": "54462360-67e2-47e6-9962-21ba7ec7f141",
  "target_account_id": "61b72db1-c001-4bf1-9a72-b2d6c6c8d8bd",
  "amount": 10.50,
  "currency": "EUR"
}
```
**Response:**

- Status Code: 201 OK - Transaction successful
- Status Code: 400 Bad Request - Validation error or one of the acceptance criteria failed
- Status Code: 500 Internal Server Error - An error occurred while processing the transaction

## Get all transactions
**Endpoint:** `/transactions`

**Method:** GET

**Response:**

- Status Code: 200 OK - Returns the transactions

## Get a transaction by ID
**Endpoint:** `/transactions/{id}`

**Method:** GET

**Response:**

- Status Code: 200 OK - Returns the transaction details
- Status Code: 400 Bad Request - The parameter id does not exist
- Status Code: 404 Not Found - Transaction was not found

## Get all accounts
**Endpoint:** `/accounts`

**Method:** GET

**Response:**

- Status Code: 200 OK - Returns the accounts

# Contributing
Contributions are welcome! If you have any suggestions, improvements, or bug fixes, please open an issue or submit a pull request.

# Licence
This code is licensed under the *MIT License*.

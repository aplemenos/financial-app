# financial-app
A simple microservice which handles financial transactions. The transaction service exposes a RESTful API (Application Programming Interface), allowing users to make financial transactions between two bank accounts and enforcing various acceptance criteria.

# Project Structure by feature (DDD)
The financial-app project follows the Domain-Driven Design (DDD), which it is an approach to software development that focuses on aligning the software design with the business domain, emphasizing the domain model as the central artifact of the system.
## cmd
This contains the entry point (main.go) files for all the services.
## domain
It is a pure domain package that is used by the application services. This package contains the account and transaction domains.
## aggregates
An Aggregate is a set of entities and value objects combined. In our case, the aggregates are the register and transfer application services. The register is used to create and manage an account. The transfer is used to perform a transaction from the source to the target account.
## server
It is responsible for the transport level, such as request validation, marshalling a request into an object or a struct that a service layer can interact with.
## postgres
It is the permanent store and communicates with the postgres database for storing the accounts and transactions data.
## migrations
This folder stores the schema files for creating the tables of the postgres DB. It includes also the schema for cleaning the database but in our case, we do not use it yet.
## multiplelock
It is a thread-safe map that is used to keep user’s locks. In our case, we lock the transfer critical section to block multiple access to the same account in parallel avoiding the race conditions.
## tests
It includes all integration and E2E tests
## vendor
This directory stores all the third-party dependencies locally so that the version doesn’t mismatch late

# Usage
1. Clone the repository or download the code files.
2. Open the terminal and navigate to the project directory.
3. Install the necessary dependencies by running `task tidy` and `task vendor`.
4. Run `task build` to compile the packages and their dependencies.
5. Run the service through `task run`. The service runs through a docker container.

NOTE: You should install the task runner.

# Testing
To run the unit tests for the financial microservice, execute the following command:
```
task test
```
To run the integration tests:
```
task integration-test
```
To run the E2E tests:
```
task acceptance-test
```

# Contributing
Contributions are welcome! If you have any suggestions, improvements, or bug fixes, please open an issue or submit a pull request.

# Licence
This code is licensed under the *MIT License*.

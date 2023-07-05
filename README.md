# financial-app
A simple microservice which handles financial transactions. The transaction service exposes a RESTful API (Application Programming Interface), allowing users to make financial transactions between two bank accounts and enforcing various acceptance criteria.

# Code Structure
## cmd
This will contain the entry point (main.go) files for all the services.
## pkg
The "pkg" folder will have all the code that could be used in other packages. In the current case, there is subfolder called "models", which will have the domain definitions for the account and transaction data models. That is all the struct and interface definitions, which will describe the data entities.
## internal
All the private code, which cannot be imported into other applications, will go into the internal package. Each application will have a separate subfolder within it. There's a subfolder for the `transaction` application. In the current case, it has multiple subfolders, which include:
- transport/http: This layer is responsible for the transport level, such as request validation, marshalling a request into an object or a struct that a service layer can interact with.
- service: Any specific business logic required for the application. In our case, we locate the transaction service.
- database: It communicates with the postgres database for storing the accounts and transactions data.
## migrations
This folder stores the schema files for creating the tables of the postgres DB. It includes also the schema for cleaning the database but in our case, we do not use it yet.
## tests
It will have all integration and E2E tests
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

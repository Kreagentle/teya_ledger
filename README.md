# Building a Ledger

## Description

This project presents a ledger designed to store transaction history and account balances. To begin using the ledger, users must create an account. For each account, the following functionalities are supported: recording new transactions, viewing transaction history, and checking the current account balance.

## How to run the solution
Ensure you are in the teya_ledger directory.

### Prerequirenments - Docker
Follow these steps to build and run application using Docker (it includes tests as well):
```bash
docker build -t ledger_app . && docker run -p 8000:8000 ledger_app
```

### Prerequirenments - go version 1.23

How to run application locally:
```bash
go run ./cmd/ledger-server
```

How to run tests locally:
```bash
go test ./...
```

## What has been accomplished

- **Assumptions** <br>

The ability to have multiple accounts has been added, allowing several users to exist within the same system. To start using the system, you need to create an account by sending a POST request to http://localhost:8000/account. This will return an account ID, which should be used in subsequent requests.

- **Ability to record money movements (ie: deposits and withdrawals)**<br>

To add funds to an account, a deposit must be made. To do this, send a transaction (POST http://localhost:8000//accounts/:id/transactions) with a request body specifying the Type (currently, only "deposit" and "withdraw" are supported) and the Amount. The response will include the updated account balance along with the transaction details, including the input parameters, transaction ID, and the time the transaction was made. The transaction status is determined by the HTTP code: if it is 200, the transaction was successful. Currently, it is not possible to go into a negative balance, and if this is attempted, the response will indicate that the transaction cannot be completed.

- **View current balance**

There is an option to check the current account balance by sending a GET request to GET http://localhost:8000/balance/:id, where :id is the account ID.

- **View transaction history**<br>

It is also possible to view the transaction history for each user. To do this, you can send a GET request to http://localhost:8000/history, which will return all transactions in chronological order (with the earliest transactions appearing first). Additionally, you can use a query parameter period, which specifies the time range from the start for which transactions should be returned. The period should be in the standard format accepted by the time.Duration() function.

## Areas for improvement

When setting up the Ledger API, it's essential to prioritise scalability, security, system settings configuration, and clear documentation for future improvements.

From a functionality perspective, currently, withdrawals can only occur if the account balance is positive. It would be beneficial to add the option for negative balances, allowing users to borrow a certain amount. Additionally, when retrieving transaction history, it is currently limited to a specific time range from the beginning. It would be useful to allow the user to specify a full time range, such as from one hour ago to half an hour ago. Furthermore, transactions should be sortable and returned in the order specified by the user.

From a scalability standpoint, adding more fields, such as user details, to transaction data would be advantageous. It would also make sense to add more endpoints, for example, to retrieve not just the balance but the full account details. Additionally, future development could focus on enabling communication between accounts, allowing them to transfer money to one another.

For system settings configuration, it's important to implement logging and metrics tracking to monitor when the API experiences downtime and to investigate the causes. Increasing test coverage is also essential. Moreover, a configuration file should be added to allow for easy modification of application parameters.

In terms of security, transitioning from HTTP to HTTPS is crucial for securely transmitting transaction details. Furthermore, implementing user authentication for those uploading data to the ledger would greatly enhance security.

For documentation, providing more detailed and interactive documentation, such as using Swagger, would improve clarity and usability, ensuring that users can fully understand and interact with the API.

## Examples

### Create a New Ledger Account
Creates a new ledger account with a unique ID and initialises it with a balance of 0 and an empty transaction history.

- **Endpoint**: `POST /account`
- **Request Body**: None
- **Response**:
    - **Success (HTTP 201)**:
      ```json
      {
        "id": "1f17d3aa-e957-11ef-b6cc-a24db5acd52b"
      }
      ```
    - **Error (HTTP 500)**:
      ```json
      {
        "error": "Failed to generate account id",
        "details": "error details"
      }
      ```

---

### Get Account Balance
Retrieves the current balance of a specific ledger account.

- **Endpoint**: `GET /balance/:id`
- **Parameters**:
    - `id` The uuid of the ledger account.
- **Response**:
    - **Success (HTTP 200)**:
      ```json
      {
        "balance": "50.00"
      }
      ```
    - **Error (HTTP 400)**:
      ```json
      {
        "error": "Invalid account ID format",
        "details": "error details"
      }
      ```
    - **Error (HTTP 404)**:
      ```json
      {
        "error": "Account not found"
      }
      ```

---

### Get Transaction History
Retrieves the transaction history for a specific ledger account. Optionally, you can filter transactions by a time period.

- **Endpoint**: `GET /history/:id`
- **Parameters**:
    - `id` (URL path): The UUID of the ledger account.
    - `period` (Optional query parameter): A duration string (e.g., `24h`, `7d`) to filter transactions within the specified time period.
- **Response**:
    - **Success (HTTP 200)**:
      ```json
      [
        {
          "id": "1f17d3aa-e957-11ef-b6cc-a24db5acd521",
          "type": "deposit",
          "amount": "100.00",
          "timestamp": "2025-02-11T12:42:13Z"
        },
        {
          "id": "1f17d3aa-e957-11ef-b6cc-a24db5acd522",
          "type": "withdrawal",
          "amount": "50.00",
          "timestamp": "2025-02-11T12:43:15Z"
        }
      ]
      ```
    - **Error (HTTP 400)**:
      ```json
      {
        "error": "Invalid query parameters",
        "details": "error details"
      }
      ```
    - **Error (HTTP 404)**:
      ```json
      {
        "error": "Account not found"
      }
      ```

---

### Submit a Transaction
Submits a new transaction (deposit or withdrawal) for a specific ledger account.

- **Endpoint**: `POST /accounts/:id/transactions`
- **Parameters**:
    - `id` (URL path): The UUID of the ledger account.
- **Request Body**:
  ```json
  {
    "type": "deposit",
    "amount": "100.00"
  }
- **Response**:
    - **Success (HTTP 200)**:
      ```json
      {
        "id": "5faee446-2593-4b77-b295-3aebf8077eb6",
        "amount": "100",
        "type": "deposit",
        "timestamp": "2025-02-12T15:39:18.154435Z"
      }
      ```
    - **Error (HTTP 400)**:
      ```json
      {
        "error": "Invalid request body/Invalid account id format/Insufficient balance",
        "details": "error details"
      }
      ```
    - **Error (HTTP 404)**:
      ```json
      {
        "error": "Account not found"
      }
      ```

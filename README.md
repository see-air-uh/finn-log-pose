# finn-log-pose
A broker service that receives and forwards all incoming requests.

### Project Design
The broker is using the external package Chi for routing and middle ware.

### Project Breakdown

The broker system first authenticates all incoming requests to check if they are valid.

If they are, the request is than forwarded to the proper microservice if the incoming request is valid.


# API Docs

## Get Balance /balance/{user}

The get balance function allows you to get the balance for a user. This balance is a sum of all transaction records that exist for a user.

To utilize this function send an HTTP GET request to the server route "/balance/{user}" where {user} is the current users username. This will return a response in the format:

{
  Error bool
  Message string
  Data float32
}

where Error represents if an error has occurred,
Message represents a server message describing what has occurred in the backend,
and Data represents the current user's balance.

Currently, the getBalance function has not implemented token checks (unauthed users can check balances for all users.) Token checks will be added in the future to restrict global access.

## Update Balance
The get balance function allows you to update the balance for a user. This function will create a new transaction in the transaction table.

To utilize this function send an HTTP POST request to the server route "/balance/{user}" where {user} is the current users username, and a json BODY in the format 
{
  "transactionData": n
}

where n represents a float number. A negative number will decrement the total balance where a positive number will increment the balance. A user's balance can not fall below zero, ex: if a user has a balance of 1000, sending a value of -10000 as transactionData will throw an error.

This will return a response in the format:

{
  Error bool
  Message string
  Data float32
}

where Error represents if an error has occurred,
Message represents a server message describing what has occurred in the backend,
and Data represents the current user's new balance.

Currently, the getBalance function has not implemented token checks (unauthed users can check balances for all users.) Token checks will be added in the future to restrict global access.
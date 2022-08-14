# GCS-obi-wan
A broker service that recieves and frowards all incoming requests.

# H2 Project Design
The broker is using the external package Chi for routing and middle ware.

# H2 Project Breakdown

The broker system first authenticates all incoming requests to check if they are valid.

If they are, the request is than forwarded to the proper microservice if the incoming request is valid.

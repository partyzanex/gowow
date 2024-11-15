# gowow

Design and implement “Word of Wisdom” tcp server.

- TCP server should be protected from DDOS attacks with the Proof of Work (https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.
- The choice of the POW algorithm should be explained.
- After Proof Of Work verification, server should send one of the quotes from “word of wisdom” book or any other collection of the quotes.
- Docker file should be provided both for the server and for the client that solves the POW challenge

## Proof of Work (PoW) Overview

### How PoW Works:

- **Hash Function:** `SHA-256` hash function is used. This function produces a 256-bit hash value, 
    which then check against the difficulty requirements.

- **Difficulty:** The difficulty is determined by the number of leading zero bits that the hash must start with.

- **Goal:** The goal is to find a number called a nonce. When this `nonce` is combined with a given `prefix` and 
hashed, it should produce a hash that starts with the required number of zero bits.

- **Finding a Nonce:** To find a valid `nonce`, many possible nonce values are tried until the resulting 
hash meets the difficulty requirement (i.e., has the required number of leading zero bits).

### Steps to Perform PoW:

- **Receive a Task:** The client receives a task with two pieces of information:
    - A `prefix` (random bytes)
    - A `difficulty` level
- **Nonce Search:** The client concatenates the `prefix` with different `nonce` values and computes the hash for each attempt.
- **Difficulty Check:** For each computed hash, the client checks if it meets the difficulty requirement (the necessary number of leading zero bits).
- **Submit the Solution:** when the correct `nonce` is found, the client sends the solution back to the server for verification.

### Comparison with Hashcash

This PoW algorithm is similar to **Hashcash**, which like Bitcoin’s PoW. The main difference is that this 
algorithm uses a single SHA-256 hashing process, while Bitcoin uses a double SHA-256 hashing process for 
increased security.


### Why?

1. **Performance:** Using a single SHA-256 hashing process requires fewer computational resources, 
which speeds up finding the correct solution and reduces the overall system load. This is especially 
helpful when the server handles a large number of requests or when system resources are limited. 
While Bitcoin uses double hashing to improve security, such a high level of security is not 
necessary solely for preventing DDoS attacks.
2. **Simple Difficulty Adjustment:** This algorithm sets difficulty based on the number of leading zero
bits required in the hash. This method allows for smooth difficulty adjustments in different 
server load scenarios, making it flexible and easier to manage.

## Usage

Build project:
```shell
make build-server && ./bin/gowow-server -h
```
output:
```
NAME:
   gowow-server - A new cli application

USAGE:
   gowow-server [global options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --address value           (default: ":7700") [%GOWOW_SERVER_ADDRESS%]
   --timeout value           (default: 5s) [%GOWOW_SERVER_TIMEOUT%]
   --difficulty value        (default: 22) [%GOWOW_SERVER_DIFFICULTY%]
   --random-bytes value      (default: 8) [%GOWOW_SERVER_RANDOM_BYTES%]
   --quotes-file-path value  (default: "./assets/quotes.txt") [%GOWOW_SERVER_QUOTES_FILE_PATH%]
   --help, -h                show help

```

build client locally:
```shell   
make build-client && ./bin/gowow-client -h
```
output:
```
Usage of gowow-client:
  -address string
        the server address (default ":7700")
  -timeout duration
        the response timeout (default 5s)
```

Build docker images:
```shell
make docker-build-server
```
```shell
make docker-build-client
```

Run server:
```shell
make docker-server
```

Run client:
```shell
make docker-client
```

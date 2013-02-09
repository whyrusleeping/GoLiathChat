#Spec for Working



##Message Protocol 
###Messages
|  Flag  | Timespamp | Username Length | Username   | Message Length  |    Message   |
|:------:|:---------:|:---------------:|:----------:|:---------------:|:------------:|
| 1 byte |  4 bytes  |  2 bytes        | ulen bytes | 2 bytes         |'Length' bytes|

##Authentication Protocol

- Client connects to server via tls on port 10234
- Client sends tLogin byte (0x03)
- Server sends 32 byte 'challenge'
- Client generates 32 byte 'response'
- Client concatenates the servers challenge to its response to create a salt
- Client hashes its password using scrypt, the salt it created and the following parameters
	- 16384
	- 8
	- 1
	- and a length of 32
- The Client then sends the 32 byte hash and the 32 byte 'response'
- The Server reads each and generates a hash with the clients 'response'
- If the Server generated hash does not match the one the user sent, the server closes the connection and exits
- The server then hashes the users password again, with the same salt and different parameters
	- 32768
	- 4
	- 7
	- and a length of 32
- The server then sends the newly generated hash where the client recreates and validates it.
- If everything checks out, the authentication is complete.

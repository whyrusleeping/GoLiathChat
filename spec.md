#Spec for Working



##Message Protocol 
###Packet Format
|  Flag  | Timestamp | Username Length | Username   | Message Length  |    Message   |
|:------:|:---------:|:---------------:|:----------:|:---------------:|:------------:|
| 1 byte |  4 bytes  |  2 bytes        | ulen bytes | 4 bytes         |'Length' bytes|

###String Format
The string format used throughout the program is simply 4 bytes representing the length of the string as an int and the string as a utf8 byte array

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

##Socket API
We have written an API for the chat client so anyone can write a chat client in any language that can open a TCP socket.
First, make a TCP connection to the client (generally 127.0.0.1:10235) and send the following:
- Hostname of the server
- Username
- Password
- Login Flags byte

If the authentication is successful, you can begin reading packets (see Packet Format) from the socket, and writing string format messages to it. 

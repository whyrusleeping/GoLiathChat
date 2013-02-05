#Spec for Working



##Message Protocol 
###Messages
|  Flag  | Timespamp |  Length  |    Message   |
|:------:|:---------:|:--------:|:------------:|
| 1 byte |  4 bytes  |  2 bytes |'Length' bytes|

##Authentication Protocol

- Client connections and requests authentication
- Server sends client its public key
- client then uses the servers public key to encrypt username and password
- client then sends server its public key
- messages that the client sends the server are encrypted with the servers public key
- messages that the server sends the clients are encrypted with the clients public key

#Go Commandline Chat

Command line chat is a simple server-client chat system with 
login and verification implemented in go.

We use termbox-go, a curses like command line window manager
 to smooth the user interface

We use port 10234 to communicate, be sure to have it open.

##Goals
Our main goal for this project is to have a quick and secure chatroom that will be as simple as possible to set up and connect to.

##Server
Features:



##Client
**Login**
<br>We allow for authentication through the `login` command

      /login [username] [password]
leaving the password space blank attempts to login without verification

**History**
<br>We allows users to request history through using the `history` command

      /history [time]

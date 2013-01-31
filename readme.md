#Go Commandline Chat

Command line chat is a simple server-clent chat system with 
login and verification implemented in go.

We use termbox-go, a curses like command line window manager
 to smooth the user interface

We use port **PORT HERE** to communicate, be sure to have it open.

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

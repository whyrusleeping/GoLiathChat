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
The first screen that comes up will be the login screen, enter your username and password to authenticate with the server.
If you do not have a username and password on the server, press ctrl+r to bring up the registration screen.

**Registration**
From the registration screen, enter your desired username and password and submit it. Your request gets sent to the server where it is reviewed by moderators and accepted or denied. If it is accepted you may then log into the server with your username and password.

**History**
<br>We allows users to request history through using the `history` command

      /history [number of messages]

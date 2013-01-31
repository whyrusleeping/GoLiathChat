#Spec for Working



##Message Protocol 
###Messages
|  Flag  | Timespamp |  Length  |    Message  |
|:------:|:---------:|:--------:|:------------:|
| 1 byte |  4 bytes  |  2 bytes | 2^16-1 bytes |

###Command
|  Flag  | Timespamp |  Command  |  Arg Length  |  Arguements  |
|:------:|:---------:|:---------:|:------------:|:------------:|
| 1 byte | 4 bytes   |  1 byte   |    2 bytes   | 2^16-1 bytes |





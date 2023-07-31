# Turn Server

1. create `secret/proxy.json` file with proxy account: 
```
    {"username":"your_proxy_email","password":"your_proxy_pwd"}
```
2. open `service.sh`, change path of `ExecStart`, `WorkingDirectory` to current directory
```
    ExecStart           = /home/my_linux/turn/turn
    WorkingDirectory    = /home/my_linux/turn
```
3. deploy turn server
```
    $ sudo sh service.sh
```
---
Notice
- Your server deploy turn need to open port from  49152 to 65535
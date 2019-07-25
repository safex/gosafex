* API *
** Create wallet file **
- URL: /init/create
- HTTP METHOD: POST
- REQUEST:
    * path - string - Path where wallet file is being created
    * password - string - Password used for encryption of wallet file
    * nettype - string - Which type of network are we going to connect

 - EXAMPLE REQUEST: 
   ``` {
        "path" : "/home/stefan/workspace/test_go_wallet/blah.bin",
        "password" : "x",
        "nettype" : "mainnet"
    }```
 - RESPONSE:
    * accounts - array of strings - Representing list of accounts being in wallet.
 - EXAMPLE RESPONSE: 
   ```
   {
    "result": {
        "accounts": [
            "primary"
        ]
    },
    "status": 0,
    "JSONRpcVersion": "1.0.0"
   }
   ``` 

** Open wallet file **
- URL: /init/open
- HTTP METHOD: POST
- REQUEST:
    * path - string - Path where wallet file is located
    * password - string - Password used for encryption of wallet file
    * nettype - string - Which type of network are we going to connect

 - EXAMPLE REQUEST: 
   ``` {
        "path" : "/home/stefan/workspace/test_go_wallet/blah.bin",
        "password" : "x",
        "nettype" : "mainnet"
    }```
 - RESPONSE:
    * accounts - array of strings - Representing list of accounts being in wallet.
 - EXAMPLE RESPONSE: 
   ```
   {
    "result": {
        "accounts": [
            "primary"
        ]
    },
    "status": 0,
    "JSONRpcVersion": "1.0.0"
   }
   ``` 
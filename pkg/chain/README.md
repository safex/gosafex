# Quick FileWallet reference

## Local structure

* One file can hold multiple wallets, each wallet is identified by a name and encrypted with a masterpass
* Wallets are kept logically separated within the file
* Every wallet name, key and value is encoded with the masterpass and nonces, to ensure data privacy

## Wallet structure

* Relevant wallet keys are
    1. WalletInfoKey - Contains wallet name and other info
    2. BlockReferenceKey - Contains a list of known block IDs 
    3. LastBlockReferenceKey - Contains the ID of the last know block
    4. OutputReferenceKey - Contains a list known output IDs
    5. OutputTypeReferenceKey - Contains a list of known output types
    6. UnspentOutputReferenceKey - Contains a list of known unspent output IDs
    7. TransactionInfoReferenceKey - Contains a list of known Transaction References


    * Single object keys:
        1. "**Out-**"    + **outputID**     - Single output saved as marshalled protobuf  
        2. "**OutInfo-** + **outputID**     - Single OutputInfo 
        3. "**Blk-**"    + **blockID**      - Single block header
        4. "**Typ-**"    + **outputType**   - List of output IDs referring to outputs of the given type
        5. "**TxInfo-**" + **transactionID**- Single TransactionInfo
        6. "**Txs-**"    + **blockID**      - List of transactions contained in the block
        7. "**TxOuts-**" + **transactionID**- List of output IDs referring to outputs contained in the tx

* How are **_IDs_** calculated
    * **outputID**      - **byte(** blockHash **)** + **byte(** output.LocalIndex **)** ; 
    * **blockID**       - **byte(** blockHash **)**
    * **outputType**    - **string**
    * **transactionID** - **string**

* Custom **Types** utilized in the wallet 
    * **OutputInfo** Contains filewallet data relative to a given output:
        1. outputType       - One of the outputTypes contained in the wallet
        2. blockHash        - The blockHash relative to the origin of the output  
        3. transactionID    - The transactionID relative to the origine of the output
        4. txLocked         - The lock status of the transaction, expressed as a single char, "U" or "L"
        5. txType           - The txType of the origin of this output

## Storage Example

* WalletInfoKey: DATA

* BlockReferenceKey: .....;.....;*0xa02b3f*;.....     
                                                                        
* LastBlockReferenceKey: .....                                          
                                                                        
* OutputReferenceKey: .....;**_0xa02b3f09_**;.....  
                                                                       
* OutputTypeReferenceKey: Cash;Token;.....                     
                                                                       
* UnspentOutputReferenceKey: .....;.....;**_0xa02b3f09_**   
                                                                       
* TransactionInfoReferenceKey: .....;**Jxnw2ir!ir**;.....    
                                                                          
* Blk-0xa02b3f: DATA 
                                                                      
* Txs-0xa02b3f: .....;**Jxnw2ir!ir**;.....
                                                                   
* TxInfo-Jxnw2ir!ir: DATA   
                                                                   
* TxOuts-Jxnw2ir!ir: .....;.....;.....;**_0xa02b3f09_**;.....
                                                                    
* Out-0xa02b3f09: DATA
                                                                    
* OutInfo-0xa02b3f09: DATA
                                                                    
* Typ-Cash: .....;**_0xa02b3f09_**
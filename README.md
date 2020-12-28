nano-token-cli
==============

This command-line tool allows you to create token chains and to perform operations on them. For example you can create a new token, transfer tokens and atomically swap tokens.

Install
-------

    go get -u github.com/hectorchu/nano-token-cli

How does it work?
-----------------

Normally in Nano every account chain is asynchronous so that there is no single linear order of events. But it is possible to use a single account as a 'blockchain' which we can write messages to, and those messages will be deterministically ordered. Furthermore we can make this chain publically writeable by writing the seed to the account in the very first opening block.

To write arbitrary data within a block it is enough to set the `representative` field of the block. Any 32-byte string can be treated as a public key and converted to a Nano address, which we use to set the field with.

The token chain is built with receive blocks so that an attacker cannot perform a Denial-of-Service attack. This is because the proof-of-work required for receive blocks is much less than send blocks. In order for an attacker to cause a fork they would need to produce both a send and receive block, while a legitimate user only needs to produce a receive block.

A message is written to the token chain by producing a send block from a source account to the token chain account, with the message in the `representative` field. Then the message is 'confirmed' by producing a receive block on the token chain which links to the send block.

Some token operations need a destination account, such as the transfer operation. This is represented by a send block immediately preceding the usual send message block, where the destination of the send is the desired destination account, and the `representative` field is set to the same message as the subsequent block.

Overall, the token chain can be parsed one block at a time, processing the messages within the blocks, and discarding any invalid blocks or messages. Messages include token genesis, token transfer and token swap operations.

How do swaps work?
------------------

A swap is initiated by one side, specifying a counterparty account, the token they wish to give and the amount. The swap may then be accepted by the counterparty, which also specifies the token and amount they want to give. After this the swap can be executed with a confirm message from the initiator. At any time the swap may be cancelled by either party.

Using the tool
--------------

The tool creates a wallet with a random seed and saves the seed to `wallet.bin`. One account from the wallet may be selected at any time. To change which account is selected use the (a) option and specify the index (0 for first account, 1 for second account etc.). Every account should be funded with a minimal amount of Nano such as 0.000001. This is more than enough as messages only consume 1 Raw. To begin using a token chain you can create one with (c) or load an existing one with (l). Use (p) to parse any messages made by other users. The other options are self-explanatory.

xcbkey
======

xcbkey is a simple command-line tool for working with Core keyfiles.


# Usage

### `xcbkey generate`

Generate a new keyfile.
If you want to use an existing private key to use in the keyfile, it can be 
specified by setting `--privatekey` with the location of the file containing the 
private key.


### `xcbkey inspect <keyfile>`

Print various information about the keyfile.
Private key information can be printed by using the `--private` flag;
make sure to use this feature with great caution!


### `xcbkey signmessage <keyfile> <message/file>`

Sign the message with a keyfile.
It is possible to refer to a file containing the message.
To sign a message contained in a file, use the `--msgfile` flag.


### `xcbkey verifymessage <address> <signature> <message/file>`

Verify the signature of the message.
It is possible to refer to a file containing the message.
To sign a message contained in a file, use the --msgfile flag.


### `xcbkey changepassword <keyfile>`

Change the password of a keyfile.
use the `--newpasswordfile` to point to the new password file.


### `xcbkey generate-mnemonic`

Generate a new wallet with a BIP-39 mnemonic phrase.
The mnemonic can later be used to recover the private key.

```
xcbkey generate-mnemonic                          # 24-word mnemonic, mainnet
xcbkey generate-mnemonic --words 12               # 12-word mnemonic
xcbkey generate-mnemonic --lang japanese           # Japanese wordlist
xcbkey generate-mnemonic --networkid 3             # Devin testnet (ab prefix)
xcbkey generate-mnemonic --account 1               # Derive account 1
```

Use `--account N` to derive multiple wallets from the same mnemonic.


### `xcbkey mnemonic-to-key <mnemonic words...>`

Recover a private key from a BIP-39 mnemonic phrase.

```
xcbkey mnemonic-to-key word1 word2 ... word24
xcbkey mnemonic-to-key --account 1 word1 word2 ... word24
xcbkey mnemonic-to-key --networkid 3 word1 word2 ... word24
```


### `xcbkey verify-mnemonic <mnemonic words...>`

Verify that a mnemonic derives the expected private key.
Requires either `--privatekey <hex>` or `--keyfile <path>`.

```
xcbkey verify-mnemonic --privatekey <hex> word1 word2 ... word24
xcbkey verify-mnemonic --keyfile /path/to/key word1 word2 ... word24
xcbkey verify-mnemonic --account 3 --privatekey <hex> word1 word2 ... word24
```


## Passwords

For every command that uses a keyfile, you will be prompted to provide the 
password for decrypting the keyfile.  To avoid this message, it is possible
to pass the password by using the `--passwordfile` flag pointing to a file that
contains the password.

## JSON

In case you need to output the result in a JSON format, you shall by using the `--json` flag.

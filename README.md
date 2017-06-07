# piratcloud
an ipfs-based encrypted backup solution that lets friends keep backup of each others' important stuff



## How it works
It basically compresses the target file/folder using tar, encrypts the tarball and uploads the encrypted tarball to ipfs. 

When you download and restore a backup the process is run in reverse.

Every time you upload or rehost something, that fact is saved in a flatfile database (a json file). The flatfile database with your uploaded files and their decryption keys exist at `~/.config/piratcloud`.

Since the tarball is encrypted, friends that are rehosting your hash can't read its contents. Which is great!
If other computers are rehosting your hash, all you need to do is keep a backup of `~/.config/piratcloud`!

## Usage 

### Backup a folder
```sh
cloud upload <directory|file> [optional note to remember what it was]

e.g.
cloud upload ~/.config # spits out the resulting ipfs hash & decryption key
```



### Rehost someone else's stuff
```sh
cloud rehost <ipfs hash> [optional note to remember why you are rehosting this]

e.g.
cloud rehost Qm....7331 "best friend backup" # Qm...7331 being the ipfs hash they give you
```



### Download your stuff
```sh
cloud download <desination dir> <ipfs hash> <decryption key>

e.g.
cloud download ~/destination-folder Qm....7331 D3crYpt100nc3i # Qm...7331 being the ipfs hash they give you
``` 



### List all your uploads and rehosts
```sh
cloud list
```


### Full command list
```
Usage:
  cloud [command]

Available Commands:
  download    Downloads an ipfs hash and decrypts it using the supplied key
  help        Help about any command
  list        Lists the stuff you've uploaded, their keys and also what you're rehosting
  rehost      Rehost an ipfs hash, basically seeding it for someone else.
  upload      Uploads and encrypts a file or directory, returning its hash and decryption key

Flags:
  -h, --help   help for cloud

Use "cloud [command] --help" for more information about a command.
```

## Is it any good?
Yeah probably

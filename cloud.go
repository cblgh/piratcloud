package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"

	"github.com/peoples-cloud/pc/crypto"
	"github.com/peoples-cloud/pc/ipfs"
	"github.com/peoples-cloud/pc/tar"
	"github.com/spf13/cobra"
)

type BackupEntry struct {
	Hash string
	Key  string
	Note string
}

var backups = make(map[string][]BackupEntry)
var basedir string
var filename = ".piratcloud"
var note string

// save upload details into the flatfile
func save() {
	data, err := json.MarshalIndent(backups, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s", basedir, filename), data, os.FileMode(0700))
	if err != nil {
		fmt.Println(err)
		return
	}
}

// load upload hashes & keys from the flatfile
func load() {
	data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", basedir, filename))
	// we couldn't find the flat file, create it
	if err != nil {
		createDir()
		return
	}
	json.Unmarshal(data, &backups)
}

func createDir() {
	// try to create the base folder inside ~/.config
	if _, err := os.Stat(basedir); os.IsNotExist(err) {
		os.MkdirAll(basedir, os.FileMode(0700))
	}
	// try to create the file
	f, err := os.OpenFile(fmt.Sprintf("%s/%s", basedir, filename), os.O_CREATE, 0700)
	if err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func upload(target, note string) {
	// tar destination
	log.Println("creating tarball")
	tarball := fmt.Sprintf("%s/%s", basedir, "piratcloud.tar.gz")
	tar.Pack(target, tarball)
	// encrypt tar
	log.Println("encrypting tarball")
	key, tarball := crypto.Encrypt(tarball)
	log.Println("dest", tarball)
	log.Printf("key: %s\n", key)
	// upload to ipfs
	log.Println("uploading to ipfs")
	hash := ipfs.Add(tarball)
	log.Printf("hash: %s\n", hash)
	// save the upload details to our flat file database
	backups["backups"] = append(backups["backups"], BackupEntry{Hash: hash, Key: key, Note: note})
	save()
}

func download(dir, hash, key string) {
	log.Printf("download details \n\thash: %s\n\t key: %s\n", hash, key)
	// get from ipfs
	log.Println("downloading hash from ipfs")
	ipfs.Get(hash, dir)
	tarball := fmt.Sprintf("%s/%s", dir, hash)
	// decrypt
	log.Println("decrypting tar")
	crypto.Decrypt(tarball, key, tarball)
	// untar
	log.Println("unpacking tar")
	tar.Unpack(tarball, dir)
	log.Printf("unpacking %s into %s\n", tarball, dir)

	// remove encrypted file
	os.Remove(fmt.Sprintf("%s/%s", dir, hash))
}

func rehost(hash, note string) {
	ipfs.Pin(hash)
	backups["rehosts"] = append(backups["rehosts"], BackupEntry{Hash: hash, Key: "", Note: note})
	save()
}

// sets the base directory to ~/.config/piratcloud
func setBasedir() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	basedir = fmt.Sprintf("%s/.config/piratcloud", usr.HomeDir)
}

func checkArgsLength(n int, args []string, cmd *cobra.Command) {
	switch {
	case len(args) == 0:
		cmd.Help()
		os.Exit(0)
	case len(args) < n:
		fmt.Println(cmd.Name() + ": too few arguments")
		fmt.Println(cmd.Use)
		os.Exit(0)
	}
}

func main() {
	setBasedir()
	load()

	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "Lists the stuff you've uploaded, their keys and also what you're rehosting",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%60s\n", "UPLOADS")
			fmt.Printf("%10s %33s %56s\n", "Note", "Hash", "Decryption key")
			for _, entry := range backups["backups"] {
				fmt.Printf("%-20s %46s %46s\n", entry.Note, entry.Hash, entry.Key)
			}
			fmt.Printf("\n%60s\n", "REHOSTS")
			fmt.Printf("%10s %33s\n", "Note", "Hash")
			for _, entry := range backups["rehosts"] {
				fmt.Printf("%-20s %46s\n", entry.Note, entry.Hash)
			}
		},
	}

	var cmdUpload = &cobra.Command{
		Use:   "upload <directory> [optional note to remember what you uploaded]",
		Short: "Uploads and encrypts a file or directory, returning its hash and decryption key",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgsLength(1, args, cmd)
			switch len(args) {
			case 2: // we have a note, add it!
				upload(args[0], args[1])
			case 1:
				upload(args[0], "")
			}
		},
	}

	var cmdDownload = &cobra.Command{
		Use:   "download <destination> <ipfs hash> <decryption key>",
		Short: "Downloads an ipfs hash and decrypts it using the supplied key",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgsLength(3, args, cmd)
			dir, hash, key := args[0], args[1], args[2]
			download(dir, hash, key)
		},
	}

	var cmdRehost = &cobra.Command{
		Use:   "rehost <ipfs hash> [optional note to remember why you are rehosting this]",
		Short: "Rehost an ipfs hash, basically seeding it for someone else.",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgsLength(1, args, cmd)
			switch len(args) {
			case 2: // note time, add it!
				rehost(args[0], args[1])
			case 1:
				rehost(args[0], "")
			}
		},
	}

	var rootCmd = &cobra.Command{Use: "cloud"}
	rootCmd.AddCommand(cmdUpload, cmdDownload, cmdRehost, cmdList)
	rootCmd.Execute()
}

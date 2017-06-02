package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/peoples-cloud/pc/crypto"
	"github.com/peoples-cloud/pc/ipfs"
	"github.com/peoples-cloud/pc/tar"
)

type BackupEntry struct {
	Hash string
	Key  string
}

var backups []BackupEntry

var filepath = ".piratcloud"

func save(stuff []BackupEntry, savePath string) {
	data, err := json.MarshalIndent(stuff, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(savePath, data, os.FileMode(0777))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func load(savePath string) {
	fmt.Println("wow load stuff")
	data, err := ioutil.ReadFile(savePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	json.Unmarshal(data, &backups)
}

func upload(dir string) {
	// fmt.Println(password)
	// tar destination
	log.Println("creating tarball")
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println(curDir)
	// dir := "/home/cblgh/code/piratcloud/test"
	log.Println("backing up", dir)
	tarball := fmt.Sprintf("%s/%s", curDir, "piratcloud.tar.gz")
	tar.Pack(dir, tarball)
	// encrypt tar
	log.Println("encrypting tarball")
	key, tarball := crypto.Encrypt(tarball)
	log.Println("dest", tarball)
	log.Printf("key: %s\n", key)
	// upload to ipfs
	log.Println("uploading to ipfs")
	hash := ipfs.Add(tarball)
	log.Printf("hash: %s\n", hash)
	backups = append(backups, BackupEntry{Hash: hash, Key: key})
	save(backups, filepath)
}

func download(dir, hash, key string) {
	log.Printf("hash: %s\nkey: %s\n", hash, key)
	// get from ipfs
	log.Println("downloading program from ipfs")
	ipfs.Get(hash, dir)
	tarball := fmt.Sprintf("%s/%s", dir, hash)
	// decrypt
	log.Println("decrypting tar")
	crypto.Decrypt(tarball, key, tarball)
	// untar
	log.Println("unpacking tar")
	tar.Unpack(tarball, dir)
}

func main() {
	load(filepath)
	if os.Args[1] == "upload" {
		fmt.Println("upload!")
		upload(os.Args[2])
	} else if os.Args[1] == "download" {
		fmt.Println("download!!")
		dir, hash, key := os.Args[2], os.Args[3], os.Args[4]
		download(dir, hash, key)
	} else if os.Args[1] == "rehost" {
		ipfs.Pin(os.Args[2])
	}
}

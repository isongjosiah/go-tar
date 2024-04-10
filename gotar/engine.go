package gotar

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/isongjosiah/hack/tar/constants"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// TarEngine is an instance that embodies the functionality of tar
type TarEngine struct {

	// Flags contains the flag defined upon initiation.
	Flags TarFlag

	// Args
	Args []string
}

type TarFileHeader struct {

	// Name is the file name
	Name string

	// Mode is the file mode (octal)
	Mode string

	// UId is the owners numeric user ID
	UId string

	// GId is the group numeric user ID
	GId string

	// Size if the file size in bytes (octal)
	Size string

	// MTime is the last modification time
	MTime string

	// ChkSum is the check sum for header record
	ChkSum string

	// Type is the link indicator (file type)
	Type string

	// LnkName is the name of the linked file
	LnkName string

	// Magic is the Ustar indicator is one of "ustar" or NUL
	Magic string

	// Version is the ustar version
	Version string

	// UName is the owner user name
	UName string

	// GName is the owner group name
	GName string

	// DevMajor is the device major number
	DevMajor string

	// DevMinor is the device minor number
	DevMinor string

	// Prefix is the filename prefix
	Prefix string

	Content string
}

func IgniteTarEngine(flag TarFlag, args []string) *TarEngine {
	return &TarEngine{
		Flags: flag,
		Args:  args,
	}
}

func (gt *TarEngine) Execute() {

	switch {

	case gt.Flags.ListArchivedContent:

		// if the f flag is specified
		if gt.Flags.UseFile {

			fileName := gt.Args[len(gt.Args)-1] // if we are listing the content of an archive the file name would be the last
			file, err := os.Open(fileName)
			if err != nil {
				if os.IsNotExist(err) {
					log.Fatalf("tarball %s does not exist", fileName)
				}

				log.Fatalf("there was an error opening file %s: %v", fileName, err)
			}

			fileHeaders := parseTarFileHeader(file)
			for _, header := range fileHeaders {
				fmt.Println(header.Name)
			}

			return

		}

		// otherwise process content from the standard input
		scanner := bufio.NewScanner(os.Stdin)
		validated := false
		for scanner.Scan() {

			// validate that the content is a tarball
			if !validated {

				sig := scanner.Bytes()
				validator := bytes.NewReader(sig)
				version := make([]byte, 6)
				_, err := validator.ReadAt(version, 257) // offset was obtained from the ustar tar format documentation on wikipedia
				if err != nil {
					if errors.Is(err, io.EOF) {
						log.Fatalf("gotar: Error opening archive: Unrecognized archive format")
					}
					log.Fatalf("there was an error reading version: %v", err)
				}

				versionS := strings.TrimSpace(string(version))

				if !strings.Contains(versionS, "ustar") {
					log.Fatalf("gotar: Error opening archive: Unrecognized archive format")
				} else {
					validated = true
				}

			}

			text := scanner.Text()
			fName := strings.Split(text, " ")[0]
			fmt.Println(fName[:len(fName)-6])
		}

	case gt.Flags.ExtractFromArchive:

		// todo(josiah): handle newline bug
		// todo: handle input from standard in
		if gt.Flags.UseFile {

			fileName := gt.Args[len(gt.Args)-1] // if we are listing the content of an archive the file name would be the last
			file, err := os.Open(fileName)
			if err != nil {
				if os.IsNotExist(err) {
					log.Fatalf("tarball %s does not exist", fileName)
				}

				log.Fatalf("there was an error opening file %s: %v", fileName, err)
			}

			fileHeaders := parseTarFileHeader(file)

			var wg sync.WaitGroup
			for _, header := range fileHeaders {

				wg.Add(1)
				go func(fHeader TarFileHeader) {

					defer wg.Done()
					file, err = os.Create(fHeader.Name)
					if err != nil {
						fmt.Println("error:", err)
						log.Printf("gotar: Unable to extract content for %s: failed to create file\n", fHeader.Name)
						return
					}

					_, err = file.WriteString(fHeader.Content)
					if err != nil {
						log.Printf("gotar: Unable to extract content for %s: failed to write to file\n", fHeader.Name)
						return
					}

					log.Println("gotar: Extracted content for ", fHeader.Name)

				}(header)

			}
			wg.Wait()
		}

	}
}

// read reads the byte content of a file from offset up till the byteSize
// and parses it as a string
func read(file *os.File, byteSize int, offset *int) string {

	b := make([]byte, byteSize)
	blockRead, err := file.ReadAt(b, int64(*offset))
	if err != nil {
		if errors.Is(err, io.EOF) {
			return ""
		}
		log.Fatal()
	}

	if blockRead != byteSize {
		log.Fatal()
	}

	*offset += byteSize

	out := string(b)
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, out)

}

/*
parseTarFileHeader takes in a tar file processes it and parse important information
from the file header to
*/
func parseTarFileHeader(file *os.File) []TarFileHeader {

	fileHeaders := make([]TarFileHeader, 0)
	offSet := 0
	fileStat, err := file.Stat()
	if err != nil {
		log.Fatalf("could not stat file %s: %v", file.Name(), err)
	}

	for {

		// check if we are at the end of the file
		if fileStat.Size() <= int64(offSet) {
			break
		}

		// parse header content
		tempFile := TarFileHeader{
			Name:     read(file, constants.NameByteSize, &offSet),
			Mode:     read(file, constants.ModeByteSize, &offSet),
			UId:      read(file, constants.UIdByteSize, &offSet),
			GId:      read(file, constants.GIdByteSize, &offSet),
			Size:     read(file, constants.FileSizeByteSize, &offSet),
			MTime:    read(file, constants.MTypeByteSize, &offSet),
			ChkSum:   read(file, constants.ChkSumByteSize, &offSet),
			Type:     read(file, constants.TypeFlgByteSize, &offSet),
			LnkName:  read(file, constants.LnkNameByteSize, &offSet),
			Magic:    read(file, constants.MagicBytesize, &offSet),
			Version:  read(file, constants.VersionByteSize, &offSet),
			UName:    read(file, constants.UNameByteSize, &offSet),
			GName:    read(file, constants.GNameByteSize, &offSet),
			DevMajor: read(file, constants.DevMajorByteSize, &offSet),
			DevMinor: read(file, constants.DevMinorByteSize, &offSet),
			Prefix:   read(file, constants.PrefixByteSize, &offSet),
		}

		offSet += 12 // the header content parsed till this point adds up to 500 bytes - add 12 to complete the block size

		size, err := strconv.ParseInt(strings.TrimSpace(tempFile.Size), 8, 64)
		if err != nil {
			if errors.Is(err, strconv.ErrSyntax) {
				break // todo: improve this check - use io.EOF instead
			}
			fmt.Println(tempFile)
			log.Fatalf("unable to parse file size %s: %v", tempFile.Name, err)
		}

		fileSize := int(size)
		switch fileSize > constants.BlockSize {
		case true:
			fileSize += fileSize % constants.BlockSize
		case false:
			fileSize = constants.BlockSize
		}

		tempFile.Content = read(file, fileSize, &offSet)

		fileHeaders = append(fileHeaders, tempFile)

	}

	return fileHeaders
}

/*
List ...
*/
func (gt *TarEngine) List() error {
	return nil
}

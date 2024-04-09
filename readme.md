# Coding Challenge - Build Your Own Tar

This coding challenge involves building a Go CLI tool to create, list and unpack tarballs
in the Unix Standard TAR (UStar) format.

The tarball consists of a series of file objects. Each file object consists of a 512-byte header, followed by as many
512-bytes required to store the content of the file rounded to the nearest full block. Following the final file object,
there is at least two consecutive 512-bytes block of zero-filled records.

The maximum filename is 256 bytes

## Steps

1. Create test tarball using unix tar command
     ```shell
        echo "File 1 contents" >> file1.txt
        echo "File 2 contents" >> file2.txt
        echo "File 3 contents" >> file3.txt
        tar -cf files.tar file1.txt file2.txt file3.txt
     ```

2. Listing the files in the tarball using the unix tar command
   ```shell
       cat files.tar | tar -t
   ```

# cs2go

cs2go is a simple external ESP for Counter-Strike 2, written in Go.

## Building the source code

To build this project, you will need to download and install Go from the official website: https://go.dev/doc/install.

Once you have installed Go, follow these steps:

1. Download the source code and extract the ZIP file to a directory of your choice.
2. Open the directory in your terminal or command prompt.
3. Make sure your environment variables `GOOS` is set to `windows` and `GOARCH` is set to `amd64`. You can set these variables by running the following commands:
   ```
   set GOOS=windows
   set GOARCH=amd64
   ```
4. Build the project using the following command:
   ```
   go build -ldflags "-s -w"
   ```
   The `ldflags` are optional and remove debugger functionality and strip the binary resulting in smaller file size.
5. Run the program while in a match to use.

If the compiler complains about not having `"github.com/lxn/win"`, `"golang.org/x/sys/windows"`, `"golang.org/x/sys/windows"` run `go get github.com/lxn/win`, `go get golang.org/x/sys/windows`, and `go get golang.org/x/sys/windows` | Finalize with `go mod tidy`
## Example

Check out this video for an example of cs2go in action:

[![Example video](https://cdn-cf-east.streamable.com/image/0qabzg.jpg?Expires=1697594705555&Key-Pair-Id=APKAIEYUVEN4EVB2OKEQ&Signature=mbcveo4kEnn35~eiMww29PrrSFQQbxkCTh3oA9dt5LgUnTmxSMNFnK7EWJbXCJgIePjW~5Q5b85dFFMSWCFga759lRr7tJhIk46C1EmcsbhWEMcjfByJ2N7AgZGKH8RBAGsPpqBlt8cTv3pXHSRf3HeUngS7aUS~z5OZMFUfPalhY3V7LXioIbMcR85oN8-zf1vfcl3P~kZo3CYA7NhlC-YlBLjcOIlPjecj9VICa8khDhqXvwI7~hkNSJD~2P983UwPhFzfffUYt60qbOQq3wFUvv8ioEaZWRGCH6pfDYEx84ZaWnYM-sSoK3~AFIhZFhEysTE8Dhw8BFirgqpZHQ__)](https://streamable.com/0qabzg)

## Issues
Esp lagging?
Search for cs2go.exe in task manager, expand, right click the process click "go to details"
Set your priority to High, this should fix it.

Thank you for using cs2go!

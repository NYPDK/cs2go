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

If the compiler complains about not having `"github.com/lxn/win"` and/or `"golang.org/x/sys/windows"` run `go get github.com/lxn/win` and `go get golang.org/x/sys/windows` | Finalize with `go mod tidy`
## Example

Check out this video for an example of cs2go in action:

[![Example video](https://cdn-cf-east.streamable.com/image/pwd7bt.jpg?Expires=1697333786148&Key-Pair-Id=APKAIEYUVEN4EVB2OKEQ&Signature=aX~z1QiaZUgVA46Pmw-1H22cc~BM4dEtx6U~jJj0HB1bP-PzIQARLF~RtL7vhk8rXsF819C1Q2TH5IIO-g5YDhyA~gHvXE6CzONAddTsPKVXoaUhfDzbHF3JqSyTxM2AWPcA7~jjEiMnJOgg8ijSZfA4KBYEl6PbTMXj4gzawZjtly-peNil2E0akMgTZq9gJ7ev~TyQczBrddQz1pvwH7FZZY4e~HsoFQMpzpqqFYg~g7VQ~6stJy6M4mBBHe~J9k2mslpK9ZghTS4oWFy3ei372l~HgbrNTmvXUNuND~uUGKCEcdoU45FOrgJF~tDbjVwTt6nD23hkt0jpAiWYXg__)](https://streamable.com/pwd7bt)

Thank you for using cs2go!

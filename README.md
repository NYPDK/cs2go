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

## Example

Check out this video for an example of cs2go in action:

[![Example video](https://cdn-cf-east.streamable.com/image/q9pi9f.jpg?Expires=1697324454744&Key-Pair-Id=APKAIEYUVEN4EVB2OKEQ&Signature=iEE6IgMhrCNbLezEkzF6kmXbxynfLXLU2zzXLMykIM6w58lnwrbi0eF1YpsD3CAQRxVpOZsFDz-N8zWWG1JiDCOX9yeK9XbhBwUKthFafsUIzJMBDDJZ0TaC-Q36QcQ015snd8nRMH~DfS7L~p0xCdr81g~lZKDnSmU-B9qeFeW8~NdhgI0CC8JrvuzwHEp3xr44gcBNqXJzfXeiWn2MZc68UqF7sV~Vqd-8wVZySLvGnBvK2pNTa~eV02Iw-4Wrs1oWo3PBCTjyodioTBIaVT8GShyvhs~BV5PPs6PT7x1he-nUYTgtHkWOxaYjjNucZJIiE2fLtaJlYs9htTRWUA__)](https://streamable.com/q9pi9f)

Thank you for using cs2go!

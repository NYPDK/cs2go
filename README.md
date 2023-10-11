# cs2go

A simple Counter-Strike 2 external ESP written in Go.




## Building the source yourself

To build this project you will need to download Go: https://go.dev/doc/install

Download the source and extract the .zip file to wherever you like.

Open the folder location in your console
![](https://i.ibb.co/j6PRwt2/image.png)

Make sure your environment GOOS is set to Windows and GOARCH is set to amd64. To change these variables you can run `set GOOS=windows` and `set GOARCH=amd64`
![](https://i.ibb.co/qRVgV5d/image.png)

Build the project with `go build -ldflags "-s -w"` The ldflags are optional and remove debugger functionality and strip the binary resulting in smaller file size.

Run the program while in a match to use.
## Demo

Video
[![Example video](https://cdn-cf-east.streamable.com/image/q9pi9f.jpg?Expires=1697324454744&Key-Pair-Id=APKAIEYUVEN4EVB2OKEQ&Signature=iEE6IgMhrCNbLezEkzF6kmXbxynfLXLU2zzXLMykIM6w58lnwrbi0eF1YpsD3CAQRxVpOZsFDz-N8zWWG1JiDCOX9yeK9XbhBwUKthFafsUIzJMBDDJZ0TaC-Q36QcQ015snd8nRMH~DfS7L~p0xCdr81g~lZKDnSmU-B9qeFeW8~NdhgI0CC8JrvuzwHEp3xr44gcBNqXJzfXeiWn2MZc68UqF7sV~Vqd-8wVZySLvGnBvK2pNTa~eV02Iw-4Wrs1oWo3PBCTjyodioTBIaVT8GShyvhs~BV5PPs6PT7x1he-nUYTgtHkWOxaYjjNucZJIiE2fLtaJlYs9htTRWUA__)](https://streamable.com/q9pi9f)

## Installation

### Good to know for learning

If you installed a package like go get k8s.io/client-go@v0.31.1 and you want to remove it

Run:

```go get k8s.io/client-go@none```

#### Setup of the project
1. go mod init adv-go -> This will create a go.mod file in the directory.
2. specify the needed packages in the go.mod
3. Run `go mod tidy`  -> This will fetch all the packages
4. Do a docker login, if get error like this Error saving credentials: error storing credentials - err: exec: "docker-credential-desktop.exe": executable file not found in $PATH, out: ``. Refer to https://stackoverflow.com/questions/65896681/exec-docker-credential-desktop-exe-executable-file-not-found-in-path


#### Install the k8s go clients

```
go get k8s.io/client-go@v0.31.1
```

#### Run the application locally
```
go mod tidy
go run main.go
```# adv-go

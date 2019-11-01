# gRPCtemplate
Quick Start Template to begin any protobuf and gRPC server or client project

This project goes in parallel with the http://github.com/dfense/protobufModels project. It's a reference where i can keep the protobuf project separate from the main project that uses the model files. In most of my cases, i share the proto files among multiple repos, and always seem to forget the syntax on all the dependencies.

This method allows me to use go modules to make them easy to build.

# Build Instructions
to build this project, just clone the project, and run  

```go build github.com/dfense/gRPCtemplate```  

that will build the source and place the binaries in the default go directory. on osx that is: ~/go/bin

## clone and build
if you want to work on the project, it's best to clone it, and then build it like so

```
  git clone github.com/dfense/gRPCtemplate
  go build github.com/dfense/gRPCtemplate
```

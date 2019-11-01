# gRPCtemplate
Quick Start Template to begin any protobuf and gRPC server or client project

This project goes in parallel with the http://github.com/dfense/protobufModel project. It's a reference where i can keep the protobuf project separate from the main project that uses the model files. In most of my cases, i share the proto files among multiple repos, and always seem to forget the syntax on all the dependencies.

This method allows me to use go modules to make them easy to build.

# Build Instructions
to build this project, just clone the project, and run
```go build github.com/dfense/gRPCtemplate```

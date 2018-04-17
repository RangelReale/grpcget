# grpcget

grpcget is a command-line client for gRPC servers with reflection enabled.

With this tool you can query the server for services and symbols, and invoke methods with parameters.

The parameters for a method invocation are simple name=value parameters, and the name can have "." to set values in inner messages.

### example

List all services:

    grpcget -plaintext list localhost:11300

List service methods:

    grpcget -plaintext list localhost:11300 helloworld.Greeter 

Describe symbol:

    grpcget -plaintext desc localhost:11300 helloworld.Greeter 
    grpcget -plaintext desc localhost:11300 helloworld.HelloRequest 

Invoke:

    grpcget -plaintext invoke localhost:11300 helloworld.Greeter name="MyName"
    
### library

grpcget is also a customizable library that you can use in your projects.

It supports customizing setters and getters so you can define special handling for types that your application supports.

For example, give an UUID value type:

```proto
message UUID {
    string value = 1;
}
```

To invoke a parameter with this type you would need to call:

    grpcget -plaintext invoke localhost:11300 app.MyService id.value="6708164e-2a56-4312-a66c-8f4de3b7b261"

You can create a field getter that allows you to omit the ".value" part:

    grpcget -plaintext invoke localhost:11300 app.MyService id="6708164e-2a56-4312-a66c-8f4de3b7b261"

Set "DynMsgHelper" for details. 
    
### author

Rangel Reale (rangelspam@gmail.com)

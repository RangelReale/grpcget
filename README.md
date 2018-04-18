# grpcget

[![GoDoc](https://godoc.org/github.com/RangelReale/grpcget?status.svg)](http://godoc.org/github.com/RangelReale/grpcget)

grpcget is a command-line client for gRPC servers with reflection enabled.

With this tool you can query the server for services and symbols, and invoke methods with parameters.

The parameters for a method invocation are simple name=value parameters, and the name can have "." to set values in inner messages.

The default implementation output is aimed to be user-friendly, not JSON or marchine parseable.
But using output customizers it is easy to create a version that does output other formats. 

### example

List all services:

    grpcget -plaintext list localhost:11300

List service methods:

    grpcget -plaintext list localhost:11300 helloworld.Greeter 

Describe symbol:

    grpcget -plaintext describe localhost:11300 helloworld.Greeter 
    grpcget -plaintext describe localhost:11300 helloworld.HelloRequest 

Invoke:

    grpcget -plaintext invoke localhost:11300 helloworld.Greeter name="MyName"
    
### Invoke parameters

Given this protobuf message:

```proto
message SMData {
    string data = 1;
    string data2 = 2;
    map<string, SMData> data_list = 3;
    repeated SMData data_repeat = 4;
}
```
    
To set the various fields, these are sample invoke parameters:

    data="Value" 
    data_list.item1.data="Value inside the map with key 'item'" 
    data_list.item1.data2="Value inside the same map as the previous param" 
    data_repeat.0.data="Value inside the repeat with index '0'"
    data_repeat.0.data2="Value into the same index '0' as the previous"
    data_repeat.1.data2="Value inside the repeat with index '1'"
    
Notes:
* For repeated items, the index must be set in sequential order, starting with 0.
* Subsequent uses of the same map/repeated index sets the value on the same existing item.
    
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
    
### acknowledgement

This library is heavily based on [grpccurl](https://github.com/fullstorydev/grpcurl), and the packages it uses.    
    
### author

Rangel Reale (rangelspam@gmail.com)

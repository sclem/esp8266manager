
### ESP8266 Manager

What is the ESP8266?

[https://en.wikipedia.org/wiki/ESP8266](https://en.wikipedia.org/wiki/ESP8266)

* Adds a REST proxy to normal TCP commands for one or many ESP8266 modules in a network.

* Primary use case would be to set up all ESP8266 modules with known IP addresses and configure them in modules.json file. The server will maintain connection to the modules.

* Client applications will use REST to trigger commands or retrieve information about each module the server knows about.

##### Getting started

1. Clone the repo, go install or go run server/server.go

2. If you don't have any chips on hand, the fakeesp8266 package listens on port 9999 for testing.

An example module config:

````
[{
    "name": "garage door",
    "target": "192.168.1.150",
    "commands": {
        "open": {
            value: 10
        },
        "close": {
            value: 20
        }
    }
}]
````

Commands can be nested with subroutines and delays (measured in milliseconds):
(Subroutines will evaluate top to bottom)

````
[{
    "name": "kitchen lights",
    "target": "192.168.1.120",
    "commands": {
        "on": {
            "commands": [{
                value: 1
            }, {
                value: 2,
                delay: 500
            }, {
                "commands": [{
                    value: 3
                }]
            }]
        },
        "off": {
            value: 0
        }
    }
}]
````


The only restriction is unique names for module names.

API:

From any REST client:

### GET /
    - returns the list of modules

### GET /[moduleName]
    - returns a module and all of its commands, including active state (whether the server can currently communicate with it)

### GET /[moduleName]/[commandName]
    - performs a command by name

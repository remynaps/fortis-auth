# fortis

![Logo](/docs/fortis-banner-2.png "Gilden logo")

user authorization service written in go

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisities

Fortis has a number of dependencies. However, this is all taken care of using docker.
#### On OSX

1. install Go
```
brew install Go
```

### Installing

A step by step series of examples that tell you how to get a development env running.

First. Clone the repository into your preferred folder.
```
git clone https://gitlab.com/gilden/fortis.git
```

2. Build the project

   This will build the project and compile it into the throne executable. which will be placed in the same directory as the build script.

```
sh build.sh
```

3. Run Fortis

```
./fortis
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.
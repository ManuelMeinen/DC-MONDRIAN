# Gateway TP
## How to run it:
To start the Gateway TP according to `config/config.json` simply run ```go run main.go``` or ```go build main.go``` followed by ```./main```. In order to update the `config.json` file run

 ```python3 create_config.py <Hostname> <TP-Address>```
## How to run the Microbenchmarks
To reproduce the results from microbenchmarking consult the file [_cmd.txt](../Benchmarking/_cmd.txt). The individual test functions can be found in the file [microbenchmark_test.go](microbenchmark_test.go).
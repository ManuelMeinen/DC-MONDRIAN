# Topologies
## How to run the Endpoint TP Testbed
To run the Endpoint TP testbed, simply run ```sudo python3 EndpointTP_testbed.py```.
## How to run the Gateway TP Testbed
To run the Gateway TP Testbed, run ```sudo ./build_and_run.sh```. This stops and removes existing Gateway TP containers, builds the new Gateway TP Docker image containing the current code base and starts the Gateway TP testbed.
## How to run the MONDRIAN Testbed
To run the MONDRIAN testbed, run ```sudo ./clean_and_run_Mondrian_testbed.sh```. This stops and removes running Gateway TP instances and executes the MONDRIAN testbed.
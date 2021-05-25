# Endpoint TP
## How to run it:
To start the Endpoint TP according to `config/config.json` simply run ```python3 start_custom_Endpoint_TP.py```. In order to update a configuration property simply append `--key value`. This results for example in the following command:
```python3 start_custom_Endpoint_TP.py --endpointTPPort 6633 --tpAddr 1.2.3.4```
## What gets started:
Running the aforementioned command starts the `ryu-manager` on the predefined port (`endpointTPPort`), then starts the implementation of the Endpoint TP which can be found in `main.py`. Furthermore, other ryu services get started, which take care of routing/forwarding.
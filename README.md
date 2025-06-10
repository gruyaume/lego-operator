# LEGO operator

This charm is a Proof of Concept for validating the [goops](https://github.com/gruyaume/goops) library. The official charm can be found [here](https://github.com/canonical/lego-operator).

## Usage

Deploy and integrate LEGO with a TLS Certificates requirer:

```shell
juju deploy ./lego_amd64.charm
juju deploy tls-certificates-requirer
juju integrate lego tls-certificates-requirer
```

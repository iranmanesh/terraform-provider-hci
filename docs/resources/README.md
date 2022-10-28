# hypertec.cloud Provider

The hypertec.cloud provider is used to interact with the many resources supported by [hypertec.cloud](https://hypertec.cloud/). The provider needs to be configured with the proper credentials before it can be used. Optionally with a URL pointing to a running hypertec.cloud API.

In order to provide the required configuration options you need to supply the value for `api_key` field.

## Example Usage

```hcl
variable "my_api_key" {}

# Configure hypertec.cloud Provider
provider "hci" {
    api_key = "${var.my_api_key}"
}

# Create an Instance
resource "hci_instance" "instance" {
    # ...
}
```

## Argument Reference

The following arguments are supported:

- [api_key](#api_key) - (Required) This is the hypertec.cloud API key. It can also be sourced from the `HCI_API_KEY` environment variable.
- [api_url](#api_url) - (Optional) This is the hypertec.cloud API URL. It can also be sourced from the `HCI_API_URL` environment variable.

## Resources

- [**hci_environment**](environment.md)
- [**hci_instance**](instance.md)
- [**hci_load_balancer_rule**](load_balancer_rule.md)
- [**hci_network**](network.md)
- [**hci_network_acl**](network_acl.md)
- [**hci_network_acl_rule**](network_acl_rule.md)
- [**hci_port_forwarding_rule**](port_forwarding_rule.md)
- [**hci_public_ip**](public_ip.md)
- [**hci_static_nat**](static_nat.md)
- [**hci_ssh_key**](ssh_key.md)
- [**hci_volume**](volume.md)
- [**hci_vpc**](vpc.md)

# hci_instance

Create and starts an instance.

## Example Usage

```hcl
resource "hci_baremetal" "my_baremetal" {
    environment_id         = "d603a97f-f528-448a-955c-ee3e40fd45b1"
    name                   = "test-baremetal"
    network_id             = "e8360aac-cb3c-44cd-abfa-80701290e862"
    template               = "CentOS 7.9 (bare-metal)"
    compute_offering       = "metal.min.32.768.g02b"
    ssh_key_name           = "my_ssh_key"
    private_ip             = "10.2.1.124"
    dedicated_group_id      = "78fdce97-3a46-4b50-bca7-c70ef8449da8"
}
```

## Argument Reference

The following arguments are supported:

- [environment_id](#environment_id) - (Required) ID of environment
- [name](#name) - (Required) Name of baremetal instance
- [network_id](#network_id) - (Required) The ID of the network where the baremetal instance should be created
- [template](#template) - (Required) Name of template to use for the baremetal instance
- [compute_offering](#compute_offering) - (Required) Name of the compute offering to use for the baremetal instance
- [user_data](#user_data) - (Optional) User data to add to the baremetal instance
- [ssh_key_name](#ssh_key_name) - (Optional) Name of the SSH key pair to attach to the baremetal instance. Mutually exclusive with public_key.
- [public_key](#public_key) - (Optional) Public key to attach to the baremetal instance. Mutually exclusive with ssh_key_name.
- [private_ip](#private_ip) - (Optional) Instance's private IPv4 address.
- [dedicated_group_id](#dedicated_group_id) - (Optional) Dedicated group id in which the baremetal instance will be created

## Attribute Reference

In addition to the arguments listed above, the following computed attributes are returned:

- [id](#id) - ID of instance.
- [private_ip_id](#private_ip_id) - ID of baremetal's private IP
- [private_ip](#private_ip) - Baremetal's private IP

## Import

Baremetals can be imported using the baremetal id, e.g.

```bash
terraform import hci_baremetal.my_baremetal c33dc4e3-0067-4c26-a588-53c9a936b9de
```

package phpipam

import (
	"errors"
	"fmt"

	"github.com/Ouest-France/phpipam-sdk-go/controllers/subnets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePHPIPAMSubnet() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourcePHPIPAMSubnetRead,
		Schema: dataSourceSubnetSchema(),
	}
}

func dataSourcePHPIPAMSubnetRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*ProviderPHPIPAMClient).subnetsController
	out := make([]subnets.Subnet, 1)
	var err error
	// We need to determine how to get the subnet. An ID search takes priority,
	// and after that subnets.
	switch {
	case d.Get("subnet_id").(int) != 0:
		out[0], err = c.GetSubnetByID(d.Get("subnet_id").(int))
		if err != nil {
			return err
		}
	case d.Get("subnet_address").(string) != "" && d.Get("subnet_mask").(int) != 0:
		out, err = c.GetSubnetsByCIDR(fmt.Sprintf("%s/%d", d.Get("subnet_address"), d.Get("subnet_mask")))
		if err != nil {
			return err
		}
	case d.Get("section_id").(int) != 0 && (d.Get("description").(string) != "" || d.Get("description_match").(string) != "" || len(d.Get("custom_field_filter").(map[string]interface{})) > 0):
		out, err = subnetSearchInSection(d, meta)
		if err != nil {
			return err
		}
	default:
		return errors.New("No valid combination of parameters found - need one of subnet_id, subnet_address and subnet_mask, or section_id and (description|description_match|custom_field_filter)")
	}
	if len(out) != 1 {
		return errors.New("Your search returned zero or multiple results. Please correct your search and try again")
	}
	err = flattenSubnet(out[0], d)
	if err != nil {
		return err
	}
	fields, err := c.GetSubnetCustomFields(out[0].ID)
	if err != nil {
		return err
	}
	trimMap(fields)
	if err := d.Set("custom_fields", fields); err != nil {
		return err
	}
	return nil
}

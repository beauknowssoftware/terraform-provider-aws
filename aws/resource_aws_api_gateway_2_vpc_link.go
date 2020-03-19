package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsApiGateway2VpcLink() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGateway2VpcLinkCreate,
		Read:   resourceAwsApiGateway2VpcLinkRead,
		Update: resourceAwsApiGateway2VpcLinkUpdate,
		Delete: resourceAwsApiGateway2VpcLinkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsApiGateway2VpcLinkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().ApigatewayTags()

	input := &apigatewayv2.CreateVpcLinkInput{
		Name:      aws.String(d.Get("name").(string)),
		SubnetIds: expandStringList(d.Get("subnet_ids").(*schema.Set).List()),
		Tags:      tags,
	}
	if v, ok := d.GetOk("security_group_ids"); ok {
		input.SecurityGroupIds = expandStringList(v.(*schema.Set).List())
	}

	resp, err := conn.CreateVpcLink(input)
	if err != nil {
		return err
	}

	d.SetId(*resp.VpcLinkId)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{apigatewayv2.VpcLinkStatusPending},
		Target:     []string{apigatewayv2.VpcLinkStatusAvailable},
		Refresh:    apigatewayv2VpcLinkRefreshStatusFunc(conn, *resp.VpcLinkId),
		Timeout:    8 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error waiting for ApiGatewayV2 Vpc Link status to be \"%s\": %s", apigatewayv2.VpcLinkStatusAvailable, err)
	}

	return resourceAwsApiGateway2VpcLinkRead(d, meta)
}

func resourceAwsApiGateway2VpcLinkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	input := &apigatewayv2.GetVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	}

	resp, err := conn.GetVpcLink(input)
	if err != nil {
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] VPC Link %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if err := d.Set("tags", keyvaluetags.ApigatewayKeyValueTags(resp.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigatewayv2",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/vpclinks/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("name", resp.Name)
	d.Set("security_group_ids", flattenStringList(resp.SecurityGroupIds))
	d.Set("subnet_ids", flattenStringList(resp.SubnetIds))
	return nil
}

func resourceAwsApiGateway2VpcLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	var input apigatewayv2.UpdateVpcLinkInput
	input.VpcLinkId = aws.String(d.Id())

	if d.HasChange("name") {
		input.Name = aws.String(d.Get("name").(string))
	}

	_, err := conn.UpdateVpcLink(&input)
	if err != nil {
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] VPC Link %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{apigatewayv2.VpcLinkStatusPending},
		Target:     []string{apigatewayv2.VpcLinkStatusAvailable},
		Refresh:    apigatewayv2VpcLinkRefreshStatusFunc(conn, d.Id()),
		Timeout:    8 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for ApiGatewayV2 Vpc Link status to be \"%s\": %s", apigatewayv2.VpcLinkStatusAvailable, err)
	}

	return resourceAwsApiGateway2VpcLinkRead(d, meta)
}

func resourceAwsApiGateway2VpcLinkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	input := &apigatewayv2.DeleteVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVpcLink(input)

	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway VPC Link (%s): %s", d.Id(), err)
	}

	if err := waitForApiGateway2VpcLinkDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for API Gateway VPC Link (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func apigatewayv2VpcLinkRefreshStatusFunc(conn *apigatewayv2.ApiGatewayV2, vl string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apigatewayv2.GetVpcLinkInput{
			VpcLinkId: aws.String(vl),
		}
		resp, err := conn.GetVpcLink(input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, *resp.VpcLinkStatus, nil
	}
}

func waitForApiGateway2VpcLinkDeletion(conn *apigatewayv2.ApiGatewayV2, vpcLinkID string) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{apigatewayv2.VpcLinkStatusPending,
			apigatewayv2.VpcLinkStatusAvailable,
			apigatewayv2.VpcLinkStatusDeleting},
		Target:     []string{""},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.GetVpcLink(&apigatewayv2.GetVpcLinkInput{
				VpcLinkId: aws.String(vpcLinkID),
			})

			if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
				return 1, "", nil
			}

			if err != nil {
				return nil, apigatewayv2.VpcLinkStatusFailed, err
			}

			return resp, aws.StringValue(resp.VpcLinkStatus), nil
		},
	}

	_, err := stateConf.WaitForState()

	return err
}

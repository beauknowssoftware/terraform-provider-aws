package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// getTags is a helper to get the tags for a resource. It expects the
// tags field to be named "tags"
func getTagsApiGateway2(conn *apigatewayv2.ApiGatewayV2, d *schema.ResourceData, arn string) error {
	resp, err := conn.GetTags(&apigatewayv2.GetTagsInput{
		ResourceArn: aws.String(arn),
	})
	if err != nil {
		return err
	}

	err = d.Set("tags", keyvaluetags.Apigatewayv2KeyValueTags(resp.Tags))
	if err != nil {
		return err
	}

	return nil
}

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsApiGateway2(conn *apigatewayv2.ApiGatewayV2, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Apigatewayv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return nil
}

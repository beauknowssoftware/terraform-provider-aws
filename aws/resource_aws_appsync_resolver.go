package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsAppsyncResolver() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppsyncResolverCreate,
		Read:   resourceAwsAppsyncResolverRead,
		Update: resourceAwsAppsyncResolverUpdate,
		Delete: resourceAwsAppsyncResolverDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"datasource_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`).MatchString(value) {
						errors = append(errors, fmt.Errorf("%q must match [_A-Za-z][_0-9A-Za-z]*", k))
					}
					return
				},
			},
			"type_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`).MatchString(value) {
						errors = append(errors, fmt.Errorf("%q must match [_A-Za-z][_0-9A-Za-z]*", k))
					}
					return
				},
			},
			"field_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`).MatchString(value) {
						errors = append(errors, fmt.Errorf("%q must match [_A-Za-z][_0-9A-Za-z]*", k))
					}
					return
				},
			},
			"request_mapping_template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"response_mapping_template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppsyncResolverCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	kind := "UNIT"
	input := &appsync.CreateResolverInput{
		ApiId: aws.String(d.Get("api_id").(string)),
		TypeName:  aws.String(d.Get("type_name").(string)),
		FieldName:  aws.String(d.Get("field_name").(string)),
		Kind: &kind,
		RequestMappingTemplate:  aws.String(d.Get("request_mapping_template").(string)),
		ResponseMappingTemplate:  aws.String(d.Get("response_mapping_template").(string)),
	}

	if v, ok := d.GetOk("datasource_name"); ok {
		input.DataSourceName = aws.String(v.(string))
	}

	_, err := conn.CreateResolver(input)
	if err != nil {
		return err
	}

	d.SetId(d.Get("api_id").(string) + "-" + d.Get("type_name").(string) + "-" + d.Get("field_name").(string))

	return resourceAwsAppsyncResolverRead(d, meta)
}

func resourceAwsAppsyncResolverRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	apiID, typeName, fieldName, err := decodeAppsyncResolverID(d.Id())

	if err != nil {
		return err
	}

	input := &appsync.GetResolverInput{
		ApiId: aws.String(apiID),
		TypeName:  aws.String(typeName),
		FieldName:  aws.String(fieldName),
	}

	resp, err := conn.GetResolver(input)
	if err != nil {
		if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] AppSync Resolver %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("api_id", apiID)
	d.Set("arn", resp.Resolver.ResolverArn)

	d.Set("datasource_name", resp.Resolver.DataSourceName)
	d.Set("type_name", resp.Resolver.TypeName)
	d.Set("field_name", resp.Resolver.FieldName)
	d.Set("kind", resp.Resolver.Kind)
	d.Set("request_mapping_template", resp.Resolver.RequestMappingTemplate)
	d.Set("response_mapping_template", resp.Resolver.ResponseMappingTemplate)

	return nil
}

func resourceAwsAppsyncResolverUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	apiID, typeName, fieldName, err := decodeAppsyncResolverID(d.Id())

	if err != nil {
		return err
	}

	kind := "UNIT"
	input := &appsync.UpdateResolverInput{
		ApiId: aws.String(apiID),
		TypeName:  aws.String(typeName),
		FieldName:  aws.String(fieldName),
		Kind: &kind,
		RequestMappingTemplate:  aws.String(d.Get("request_mapping_template").(string)),
		ResponseMappingTemplate:  aws.String(d.Get("response_mapping_template").(string)),
	}

	if v, ok := d.GetOk("datasource_name"); ok {
		input.DataSourceName = aws.String(v.(string))
	}

	_, err = conn.UpdateResolver(input)
	if err != nil {
		return err
	}
	return resourceAwsAppsyncResolverRead(d, meta)
}

func resourceAwsAppsyncResolverDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	apiID, typeName, fieldName, err := decodeAppsyncResolverID(d.Id())

	if err != nil {
		return err
	}

	input := &appsync.DeleteResolverInput{
		ApiId: aws.String(apiID),
		TypeName:  aws.String(typeName),
		FieldName:  aws.String(fieldName),
	}

	_, err = conn.DeleteResolver(input)
	if err != nil {
		if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}

func decodeAppsyncResolverID(id string) (string, string, string, error) {
	idParts := strings.SplitN(id, "-", 3)
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("expected ID in format ApiID-TypeName-FieldName, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

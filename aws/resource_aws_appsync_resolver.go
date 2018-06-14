package aws

import (
	"log"

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

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"data_source_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"field_name": {
				Type:     schema.TypeString,
				Required: true,
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

	input := &appsync.CreateResolverInput{
		ApiId:                   aws.String(d.Get("api_id").(string)),
		DataSourceName:          aws.String(d.Get("data_source_name").(string)),
		TypeName:                aws.String(d.Get("type_name").(string)),
		FieldName:               aws.String(d.Get("field_name").(string)),
		RequestMappingTemplate:  aws.String(d.Get("request_mapping_template").(string)),
		ResponseMappingTemplate: aws.String(d.Get("response_mapping_template").(string)),
	}

	resp, err := conn.CreateResolver(input)
	if err != nil {
		return err
	}

	d.SetId(d.Get("api_id").(string) + "-" + d.Get("type_name").(string) + "-" + d.Get("field_name").(string))
	d.Set("arn", resp.Resolver.ResolverArn)
	return nil
}

func resourceAwsAppsyncResolverRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.GetResolverInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		TypeName:  aws.String(d.Get("type_name").(string)),
		FieldName: aws.String(d.Get("field_name").(string)),
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

	d.Set("arn", resp.Resolver.ResolverArn)
	d.Set("data_source_name", resp.Resolver.DataSourceName)
	d.Set("request_mapping_template", resp.Resolver.RequestMappingTemplate)
	d.Set("response_mapping_template", resp.Resolver.ResponseMappingTemplate)

	return nil
}

func resourceAwsAppsyncResolverUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.UpdateResolverInput{
		ApiId:                   aws.String(d.Get("api_id").(string)),
		DataSourceName:          aws.String(d.Get("data_source_name").(string)),
		TypeName:                aws.String(d.Get("type_name").(string)),
		FieldName:               aws.String(d.Get("field_name").(string)),
		RequestMappingTemplate:  aws.String(d.Get("request_mapping_template").(string)),
		ResponseMappingTemplate: aws.String(d.Get("response_mapping_template").(string)),
	}

	_, err := conn.UpdateResolver(input)
	if err != nil {
		return err
	}
	return resourceAwsAppsyncResolverRead(d, meta)
}

func resourceAwsAppsyncResolverDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.DeleteResolverInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		TypeName:  aws.String(d.Get("type_name").(string)),
		FieldName: aws.String(d.Get("field_name").(string)),
	}

	_, err := conn.DeleteResolver(input)
	if err != nil {
		if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}

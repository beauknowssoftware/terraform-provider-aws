package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/schema"
	"strings"
)

func resourceAwsAppsyncFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppsyncFunctionCreate,
		Read:   resourceAwsAppsyncFunctionRead,
		Update: resourceAwsAppsyncFunctionUpdate,
		Delete: resourceAwsAppsyncFunctionDelete,
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
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`).MatchString(value) {
						errors = append(errors, fmt.Errorf("%q must match [_A-Za-z][_0-9A-Za-z]*", k))
					}
					return
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"request_mapping_template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"response_mapping_template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"function_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppsyncFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	version := "2018-05-29"
	input := &appsync.CreateFunctionInput{
		ApiId: aws.String(d.Get("api_id").(string)),
		Name:  aws.String(d.Get("name").(string)),
		DataSourceName:  aws.String(d.Get("datasource_name").(string)),
		FunctionVersion:  aws.String(version),
		RequestMappingTemplate:  aws.String(d.Get("request_mapping_template").(string)),
		ResponseMappingTemplate:  aws.String(d.Get("response_mapping_template").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	response, err := conn.CreateFunction(input)
	if err != nil {
		return err
	}

	d.SetId(d.Get("api_id").(string) + "-" + *response.FunctionConfiguration.FunctionId)

	return resourceAwsAppsyncFunctionRead(d, meta)
}

func resourceAwsAppsyncFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	apiID, functionId, err := decodeAppsyncFunctionID(d.Id())

	if err != nil {
		return err
	}

	input := &appsync.GetFunctionInput{
		ApiId: aws.String(apiID),
		FunctionId:  aws.String(functionId),
	}

	resp, err := conn.GetFunction(input)
	if err != nil {
		if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] AppSync Function %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("api_id", apiID)
	d.Set("arn", resp.FunctionConfiguration.FunctionArn)
	d.Set("function_id", resp.FunctionConfiguration.FunctionId)

	d.Set("datasource_name", resp.FunctionConfiguration.DataSourceName)
	d.Set("name", resp.FunctionConfiguration.Name)
	d.Set("description", resp.FunctionConfiguration.Description)
	d.Set("request_mapping_template", resp.FunctionConfiguration.RequestMappingTemplate)
	d.Set("response_mapping_template", resp.FunctionConfiguration.ResponseMappingTemplate)

	return nil
}

func resourceAwsAppsyncFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	apiID, functionId, err := decodeAppsyncFunctionID(d.Id())

	if err != nil {
		return err
	}

	version := "2018-05-29"
	input := &appsync.UpdateFunctionInput{
		ApiId: aws.String(apiID),
		FunctionId: aws.String(functionId),
		Name:  aws.String(d.Get("name").(string)),
		DataSourceName:  aws.String(d.Get("datasource_name").(string)),
		FunctionVersion:  &version,
		RequestMappingTemplate:  aws.String(d.Get("request_mapping_template").(string)),
		ResponseMappingTemplate:  aws.String(d.Get("response_mapping_template").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err = conn.UpdateFunction(input)
	if err != nil {
		return err
	}
	return resourceAwsAppsyncFunctionRead(d, meta)
}

func resourceAwsAppsyncFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	apiID, functionId, err := decodeAppsyncFunctionID(d.Id())

	if err != nil {
		return err
	}

	input := &appsync.DeleteFunctionInput{
		ApiId: aws.String(apiID),
		FunctionId:  aws.String(functionId),
	}

	_, err = conn.DeleteFunction(input)
	if err != nil {
		if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}

func decodeAppsyncFunctionID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "-", 2)
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format ApiID-FunctionId, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}

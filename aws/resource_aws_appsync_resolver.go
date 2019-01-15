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
				ConflictsWith: []string{"pipeline_config"},
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
			"pipeline_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"functions": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
				ConflictsWith: []string{"datasource_name"},
			},
		},
	}
}

func resourceAwsAppsyncResolverCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.CreateResolverInput{
		ApiId: aws.String(d.Get("api_id").(string)),
		TypeName:  aws.String(d.Get("type_name").(string)),
		FieldName:  aws.String(d.Get("field_name").(string)),
		RequestMappingTemplate:  aws.String(d.Get("request_mapping_template").(string)),
		ResponseMappingTemplate:  aws.String(d.Get("response_mapping_template").(string)),
	}

	if v, ok := d.GetOk("datasource_name"); ok {
		input.DataSourceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pipeline_config"); ok {
		input.Kind = aws.String("PIPELINE")
		input.PipelineConfig = expandAppsyncPipelineConfig(v.([]interface{}))
	} else {
		input.Kind = aws.String("UNIT")
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

	if err := d.Set("pipeline_config", flattenAppsyncPipelineConfig(resp.Resolver.PipelineConfig)); err != nil {
		return fmt.Errorf("error setting pipeline_config: %s", err)
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

	input := &appsync.UpdateResolverInput{
		ApiId: aws.String(apiID),
		TypeName:  aws.String(typeName),
		FieldName:  aws.String(fieldName),
		RequestMappingTemplate:  aws.String(d.Get("request_mapping_template").(string)),
		ResponseMappingTemplate:  aws.String(d.Get("response_mapping_template").(string)),
	}

	if v, ok := d.GetOk("datasource_name"); ok {
		input.DataSourceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pipeline_config"); ok {
		input.Kind = aws.String("PIPELINE")
		input.PipelineConfig = expandAppsyncPipelineConfig(v.([]interface{}))
	} else {
		input.Kind = aws.String("UNIT")
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

func expandAppsyncPipelineConfig(l []interface{}) *appsync.PipelineConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})
	configuredFunctions := configured["functions"].([]interface{})

	functions := []*string{}
	for _, f := range configuredFunctions {
		fString := f.(string)
		functions = append(functions, &fString)
	}

	result := &appsync.PipelineConfig{
		Functions: functions,
	}

	return result
}

func flattenAppsyncPipelineConfig(config *appsync.PipelineConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	functions := []interface{} {}
	for _, f := range config.Functions {
		functions = append(functions, *f)
	}

	result := map[string]interface{}{
		"functions": functions,
	}

	return []map[string]interface{}{result}
}

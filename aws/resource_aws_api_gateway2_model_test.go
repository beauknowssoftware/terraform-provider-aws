package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGateway2Model_basic(t *testing.T) {
	resourceName := "aws_api_gateway_v2_model.test"
	rName := fmt.Sprintf("terraformtestaccapigwv2%s", acctest.RandStringFromCharSet(9, acctest.CharSetAlphaNum))

	schema := `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "ExampleModel",
  "type": "object",
  "properties": {
    "id": {
      "type": "string"
    }
  }
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2ModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2ModelConfig_basic(rName, schema),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content_type", "application/json"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schema", schema),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2ModelImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGateway2Model_AllAttributes(t *testing.T) {
	resourceName := "aws_api_gateway_v2_model.test"
	rName := fmt.Sprintf("terraformtestaccapigwv2%s", acctest.RandStringFromCharSet(9, acctest.CharSetAlphaNum))

	schema1 := `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "ExampleModel1",
  "type": "object",
  "properties": {
    "id": {
      "type": "string"
    }
  }
}
`
	schema2 := `
	{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"title": "ExampleModel",
		"type": "object",
		"properties": {
		  "ids": {
			"type": "array",
			"items":{
			  "type": "integer"
			}
		  }
		}
	  }
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2ModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2ModelConfig_allAttributes(rName, schema1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content_type", "text/x-json"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schema", schema1),
				),
			},
			{
				Config: testAccAWSAPIGateway2ModelConfig_basic(rName, schema2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content_type", "application/json"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schema", schema2),
				),
			},
			{
				Config: testAccAWSAPIGateway2ModelConfig_allAttributes(rName, schema1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content_type", "text/x-json"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schema", schema1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2ModelImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGateway2ModelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2_model" {
			continue
		}

		_, err := conn.GetModel(&apigatewayv2.GetModelInput{
			ApiId:   aws.String(rs.Primary.Attributes["api_id"]),
			ModelId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 model %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGateway2ModelExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 model ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.GetModel(&apigatewayv2.GetModelInput{
			ApiId:   aws.String(rs.Primary.Attributes["api_id"]),
			ModelId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSAPIGateway2ModelImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGateway2ModelConfig_api(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_v2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccAWSAPIGateway2ModelConfig_basic(rName, schema string) string {
	return testAccAWSAPIGateway2ModelConfig_api(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_model" "test" {
  api_id       = "${aws_api_gateway_v2_api.test.id}"
  content_type = "application/json"
  name         = %[1]q
  schema       = %[2]q
}
`, rName, schema)
}

func testAccAWSAPIGateway2ModelConfig_allAttributes(rName, schema string) string {
	return testAccAWSAPIGateway2ModelConfig_api(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_model" "test" {
  api_id       = "${aws_api_gateway_v2_api.test.id}"
  content_type = "text/x-json"
  name         = %[1]q
  description  = "test"
  schema       = %[2]q
}
`, rName, schema)
}

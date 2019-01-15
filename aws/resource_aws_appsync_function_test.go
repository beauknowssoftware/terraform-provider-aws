package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAppsyncFunction_basic(t *testing.T) {
	rName := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	resourceName := "aws_appsync_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncFunctionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncFunctionExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(fmt.Sprintf("apis/.+/functions/.+"))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "datasource_name", rName),
					resource.TestCheckResourceAttr(resourceName, "request_mapping_template", "test request"),
					resource.TestCheckResourceAttr(resourceName, "response_mapping_template", "test response"),
					resource.TestCheckResourceAttrSet(resourceName, "function_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsAppsyncFunctionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appsyncconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_function" {
			continue
		}

		apiID, functionId, err := decodeAppsyncFunctionID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &appsync.GetFunctionInput{
			ApiId: aws.String(apiID),
			FunctionId:  aws.String(functionId),
		}

		_, err = conn.GetFunction(input)
		if err != nil {
			if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsAppsyncFunctionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource has no ID: %s", name)
		}

		apiID, functionId, err := decodeAppsyncFunctionID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).appsyncconn

		input := &appsync.GetFunctionInput{
			ApiId: aws.String(apiID),
			FunctionId:  aws.String(functionId),
		}

		_, err = conn.GetFunction(input)

		return err
	}
}

func testAccAppsyncFunctionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_datasource" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  name   = %q
  type   = "NONE"
}

resource "aws_appsync_function" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  name = %q
  description = "test description"
  datasource_name = "${aws_appsync_datasource.test.name}"
  request_mapping_template = "test request"
  response_mapping_template = "test response"
}
`, rName, rName, rName)
}

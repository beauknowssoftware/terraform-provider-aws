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

func TestAccAwsAppsyncResolver_basic(t *testing.T) {
	rName := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	resourceName := "aws_appsync_resolver.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncResolverDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncResolverConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(fmt.Sprintf("apis/.+/types/Query/resolvers/test"))),
					resource.TestCheckResourceAttr(resourceName, "type_name", "Query"),
					resource.TestCheckResourceAttr(resourceName, "field_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "datasource_name", rName),
					resource.TestCheckResourceAttr(resourceName, "request_mapping_template", "test"),
					resource.TestCheckResourceAttr(resourceName, "response_mapping_template", "test"),
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

func testAccCheckAwsAppsyncResolverDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appsyncconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_resolver" {
			continue
		}

		apiID, typeName, fieldName, err := decodeAppsyncResolverID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &appsync.GetResolverInput{
			ApiId: aws.String(apiID),
			TypeName:  aws.String(typeName),
			FieldName:  aws.String(fieldName),
		}

		_, err = conn.GetResolver(input)
		if err != nil {
			if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsAppsyncResolverExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource has no ID: %s", name)
		}

		apiID, typeName, fieldName, err := decodeAppsyncResolverID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).appsyncconn

		input := &appsync.GetResolverInput{
			ApiId: aws.String(apiID),
			TypeName:  aws.String(typeName),
			FieldName:  aws.String(fieldName),
		}

		_, err = conn.GetResolver(input)

		return err
	}
}

func testAccAppsyncResolverConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  schema		      = "type Query { test:String }\nschema { query:Query }"
}

resource "aws_appsync_datasource" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  name   = %q
  type   = "NONE"
}

resource "aws_appsync_resolver" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  type_name = "Query"
  field_name = "test"
  datasource_name = "${aws_appsync_datasource.test.name}"
  request_mapping_template = "test"
  response_mapping_template = "test"
}
`, rName, rName)
}

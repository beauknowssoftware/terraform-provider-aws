package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAppsyncResolver(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncResolverDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncResolverConfig(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists("aws_appsync_resolver.test"),
					resource.TestCheckResourceAttrSet("aws_appsync_resolver.test", "arn"),
				),
			},
		},
	})
}

func TestAccAwsAppsyncResolver_update(t *testing.T) {
	rName := acctest.RandString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncResolverDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncResolverConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists("aws_appsync_resolver.test"),
					resource.TestCheckResourceAttrSet("aws_appsync_resolver.test", "arn"),
					resource.TestCheckResourceAttr("aws_appsync_resolver.test", "data_source_name", "original"),
				),
			},
			{
				Config: testAccAppsyncResolverConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists("aws_appsync_resolver.test"),
					resource.TestCheckResourceAttrSet("aws_appsync_resolver.test", "arn"),
					resource.TestCheckResourceAttr("aws_appsync_resolver.test", "data_source_name", "update"),
				),
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

		input := &appsync.GetResolverInput{
			ApiId:     aws.String(rs.Primary.Attributes["api_id"]),
			TypeName:  aws.String(rs.Primary.Attributes["type_name"]),
			FieldName: aws.String(rs.Primary.Attributes["field_name"]),
		}

		_, err := conn.GetResolver(input)
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
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAppsyncResolverConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name = "tf_appsync_%s"
}

resource "aws_appsync_schema" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  definition = <<EOF
schema {
    query:tf_appsync_%s
}

type tf_appsync_%s {
    tf_appsync_%s: [String]
}
EOF
}

resource "aws_appsync_datasource" "original" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  name = "original"
  type = "NONE"
}

resource "aws_appsync_resolver" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  field_name = "tf_appsync_%s"
  type_name = "tf_appsync_%s"
  data_source_name = "${aws_appsync_datasource.original.name}"
  request_mapping_template = "#set ($myMap = {})"
  response_mapping_template = "#set ($myMap = {})"

  depends_on = ["aws_appsync_schema.test"]
}
`, rName, rName, rName, rName, rName, rName)
}

func testAccAppsyncResolverConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name = "tf_appsync_%s"
}

resource "aws_appsync_schema" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  definition = <<EOF
schema {
    query:tf_appsync_%s
}

type tf_appsync_%s {
    tf_appsync_%s: [String]
}
EOF
}

resource "aws_appsync_datasource" "original" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  name = "original"
  type = "NONE"
}

resource "aws_appsync_datasource" "update" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  name = "update"
  type = "NONE"
}

resource "aws_appsync_resolver" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  field_name = "tf_appsync_%s"
  type_name = "tf_appsync_%s"
  data_source_name = "${aws_appsync_datasource.update.name}"
  request_mapping_template = "#set ($myMap = {})"
  response_mapping_template = "#set ($myMap = {})"

  depends_on = ["aws_appsync_schema.test"]
}
`, rName, rName, rName, rName, rName, rName)
}

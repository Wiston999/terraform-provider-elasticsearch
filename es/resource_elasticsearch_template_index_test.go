package es

import (
	"context"
	"errors"
	"fmt"
	"testing"

	elastic7 "github.com/olivere/elastic/v7"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccElasticsearchTemplateIndex(t *testing.T) {
	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}
	meta := provider.Meta()
	var allowed bool
	switch meta.(type) {
	case *elastic7.Client:
		allowed = true
	default:
		allowed = false
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("/_index_template endpoint only supported on ES >= 7.8")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchTemplateIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchTemplateIndex,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchTemplateIndexExists("elasticsearch_template_index.test"),
				),
			},
		},
	})
}

func TestAccElasticsearchTemplateIndex_importBasic(t *testing.T) {
	provider := Provider().(*schema.Provider)
	err := provider.Configure(&terraform.ResourceConfig{})
	if err != nil {
		t.Skipf("err: %s", err)
	}
	meta := provider.Meta()
	var allowed bool
	switch meta.(type) {
	case *elastic7.Client:
		allowed = true
	default:
		allowed = false
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if !allowed {
				t.Skip("/_index_template endpoint only supported on ES >= 7.8")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchTemplateIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchTemplateIndex,
			},
			{
				ResourceName:      "elasticsearch_template_index.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchTemplateIndexExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No index template ID is set")
		}

		meta := testAccProvider.Meta()

		var err error
		switch client := meta.(type) {
		case *elastic7.Client:
			_, err = client.IndexGetIndexTemplate(rs.Primary.ID).Do(context.TODO())
		default:
			err = errors.New("/_index_template endpoint only supported on ES >= 7.8")
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testCheckElasticsearchTemplateIndexDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_template_index" {
			continue
		}

		meta := testAccProvider.Meta()

		var err error
		switch client := meta.(type) {
		case *elastic7.Client:
			_, err = client.IndexGetTemplate(rs.Primary.ID).Do(context.TODO())
		default:
			err = errors.New("/_index_template endpoint only supported on ES >= 7.8")
		}

		if err != nil {
			return nil // should be not found error
		}

		return fmt.Errorf("Index template %q still exists", rs.Primary.ID)
	}

	return nil
}

var testAccElasticsearchTemplateIndex = `
resource "elasticsearch_template_index" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "index_patterns": ["te*", "bar*"],
  "template": {
    "settings": {
			"index": {
	      "number_of_shards": 1
			}
    },
    "mappings": {
      "properties": {
        "host_name": {
          "type": "keyword"
        },
        "created_at": {
          "type": "date",
          "format": "EEE MMM dd HH:mm:ss Z yyyy"
        }
      }
    },
    "aliases": {
      "mydata": { }
    }
  },
  "priority": 200,
  "version": 3
}
EOF
}
`

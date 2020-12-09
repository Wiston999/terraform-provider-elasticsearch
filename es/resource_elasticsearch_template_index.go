package es

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	elastic7 "github.com/olivere/elastic/v7"
	elastic5 "gopkg.in/olivere/elastic.v5"
	elastic6 "gopkg.in/olivere/elastic.v6"
)

func resourceElasticsearchTemplateIndex() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchTemplateIndexCreate,
		Read:   resourceElasticsearchTemplateIndexRead,
		Update: resourceElasticsearchTemplateIndexUpdate,
		Delete: resourceElasticsearchTemplateIndexDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"body": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressIndexTemplate,
				ValidateFunc:     validation.StringIsJSON,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchTemplateIndexCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceElasticsearchPutTemplateIndex(d, meta, true)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceElasticsearchTemplateIndexRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var result string
	var err error
	switch client := meta.(type) {
	case *elastic7.Client:
		version, err := elastic7GetVersion(client)
		if err == nil {
			if version < "7.8.0" {
				err = errors.New("index_template endpoint only available from ElasticSearch >= 7.8")
			} else {
				result, err = elastic7GetIndexTemplate(client, id)
			}
		}
	default:
		err = errors.New("index_template endpoint only available from ElasticSearch >= 7.8")
	}
	if err != nil {
		if elastic7.IsNotFound(err) || elastic6.IsNotFound(err) || elastic5.IsNotFound(err) {
			log.Printf("[WARN] Index template (%s) not found, removing from state", id)
			d.SetId("")
			return nil
		}

		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("name", d.Id())
	ds.set("body", result)
	return ds.err
}

func elastic7GetIndexTemplate(client *elastic7.Client, id string) (string, error) {
	res, err := client.IndexGetIndexTemplate(id).Do(context.TODO())
	if err != nil {
		return "", err
	}

	t := res.IndexTemplates[0]
	tj, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(tj), nil
}

func resourceElasticsearchTemplateIndexUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchPutTemplateIndex(d, meta, false)
}

func resourceElasticsearchTemplateIndexDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var err error
	switch client := meta.(type) {
	case *elastic7.Client:
		version, err := elastic7GetVersion(client)
		if err == nil {
			if version < "7.8.0" {
				err = errors.New("index_template endpoint only available from ElasticSearch >= 7.8")
			} else {
				err = elastic7DeleteIndexTemplate(client, id)
			}
		}
	default:
		err = errors.New("index_template endpoint only available from ElasticSearch >= 7.8")
	}

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func elastic7DeleteIndexTemplate(client *elastic7.Client, id string) error {
	_, err := client.IndexDeleteIndexTemplate(id).Do(context.TODO())
	return err
}

func resourceElasticsearchPutTemplateIndex(d *schema.ResourceData, meta interface{}, create bool) error {
	name := d.Get("name").(string)
	body := d.Get("body").(string)

	var err error
	switch client := meta.(type) {
	case *elastic7.Client:
		err = elastic7PutIndexTemplate(client, name, body, create)
	default:
		err = errors.New("index_template endpoint only available from ElasticSearch >= 7.8")
	}

	return err
}

func elastic7PutIndexTemplate(client *elastic7.Client, name string, body string, create bool) error {
	_, err := client.IndexPutIndexTemplate(name).BodyString(body).Create(create).Do(context.TODO())
	return err
}
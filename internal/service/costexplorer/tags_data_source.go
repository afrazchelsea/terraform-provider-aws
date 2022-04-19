package costexplorer

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceCostExplorerTags() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCostExplorerTagsRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     schemaCostExplorerCostCategoryRule(),
			},
			"search_string": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringLenBetween(1, 1024),
				ConflictsWith: []string{"sort_by"},
			},
			"sort_by": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"search_string"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.Metric_Values(), false),
						},
						"sort_order": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.SortOrder_Values(), false),
						},
					},
				},
			},
			"tag_key": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"time_period": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"end": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 40),
						},
						"start": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 40),
						},
					},
				},
			},
		},
	}
}

func dataSourceCostExplorerTagsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CostExplorerConn

	input := &costexplorer.GetTagsInput{
		TimePeriod: expandCostExplorerTagsTimePeriod(d.Get("time_period").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filter = expandCostExplorerCostExpressions(v.([]interface{}))[0]
	}

	if v, ok := d.GetOk("search_string"); ok {
		input.SearchString = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sort_by"); ok {
		input.SortBy = expandCostExplorerTagsSortBys(v.([]interface{}))
	}

	if v, ok := d.GetOk("tag_key"); ok {
		input.TagKey = aws.String(v.(string))
	}

	resp, err := conn.GetTagsWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error reading CostExplorer Tags (%s): %s", d.Id(), err)
	}

	d.Set("tags", flex.FlattenStringList(resp.Tags))

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return nil
}

func expandCostExplorerTagsSortBys(tfList []interface{}) []*costexplorer.SortDefinition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.SortDefinition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCostExplorerTagsSortBy(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCostExplorerTagsSortBy(tfMap map[string]interface{}) *costexplorer.SortDefinition {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.SortDefinition{}
	apiObject.Key = aws.String(tfMap["key"].(string))
	if v, ok := tfMap["sort_order"]; ok {
		apiObject.SortOrder = aws.String(v.(string))
	}

	return apiObject
}

func expandCostExplorerTagsTimePeriod(tfMap map[string]interface{}) *costexplorer.DateInterval {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.DateInterval{}
	apiObject.Start = aws.String(tfMap["start"].(string))
	apiObject.End = aws.String(tfMap["end"].(string))

	return apiObject
}

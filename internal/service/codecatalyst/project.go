package codecatalyst

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_codecatalyst_project", name="Project")
// Tagging annotations are used for "transparent tagging".
// Change the "identifierAttribute" value to the name of the attribute used in ListTags and UpdateTags calls (e.g. "arn").
// @Tags(identifierAttribute="id")
func ResourceProject() *schema.Resource {
	return &schema.Resource{

		CreateWithoutTimeout: resourceProjectCreate,
		ReadWithoutTimeout:   resourceProjectRead,
		UpdateWithoutTimeout: resourceProjectCreate,
		DeleteWithoutTimeout: resourceProjectDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"space_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameProject = "Project"
)

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).CodeCatalystClient()

	in := &codecatalyst.CreateProjectInput{
		DisplayName: aws.String(d.Get("display_name").(string)),
		SpaceName:   aws.String(d.Get("space_name").(string)),
		Description: aws.String(d.Get("description").(string)),
	}

	out, err := conn.CreateProject(ctx, in)
	if err != nil {
		return create.DiagError(names.CodeCatalyst, create.ErrActionCreating, ResNameProject, d.Get("display_name").(string), err)
	}

	if out == nil || out.Name == nil {
		return create.DiagError(names.CodeCatalyst, create.ErrActionCreating, ResNameProject, d.Get("display_name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Name))

	/* if _, err := waitProjectCreated(ctx, conn, d.Id(), *aws.String(d.Get("space_name").(string)), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.CodeCatalyst, create.ErrActionWaitingForCreation, ResNameProject, d.Id(), err)
	} */

	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).CodeCatalystClient()

	out, err := findProjectByName(ctx, conn, d.Id(), *aws.String(d.Get("space_name").(string)))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeCatalyst Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.CodeCatalyst, create.ErrActionReading, ResNameProject, d.Id(), err)
	}

	d.Set("name", out.Name)

	return nil
}

// func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// conn := meta.(*conns.AWSClient).CodeCatalystClient()

	log.Printf("[INFO] Deleting CodeCatalyst Project is not currently supprted")
	//log.Printf("[INFO] Deleting CodeCatalyst Project %s", d.Id())

	/* _, err := conn.DeleteProject(ctx, &codecatalyst.DeleteProjectInput{
		Id: aws.String(d.Id()),
	})
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.CodeCatalyst, create.ErrActionDeleting, ResNameProject, d.Id(), err)
	}


	if _, err := waitProjectDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.CodeCatalyst, create.ErrActionWaitingForDeletion, ResNameProject, d.Id(), err)
	}
	*/

	//return fmt.Errorf("CodeCatalyst Project (%s) deletion not currently supported", d.Id())
	return nil
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

/* func waitProjectCreated(ctx context.Context, conn *codecatalyst.Client, id string, space_name string, timeout time.Duration) (*codecatalyst.GetProjectOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusProject(ctx, conn, id, space_name),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*codecatalyst.GetProjectOutput); ok {
		return out, err
	}

	return nil, err
}

func waitProjectDeleted(ctx context.Context, conn *codecatalyst.Client, id string, space_name string, timeout time.Duration) (*codecatalyst.GetProjectOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusProject(ctx, conn, id, space_name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*codecatalyst.GetProjectOutput); ok {
		return out, err
	}

	return nil, err
}

func statusProject(ctx context.Context, conn *codecatalyst.Client, id string, spaceName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findProjectByName(ctx, conn, id, spaceName)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Name), nil
	}
} */

func findProjectByName(ctx context.Context, conn *codecatalyst.Client, id string, spaceName string) (*codecatalyst.GetProjectOutput, error) {
	in := &codecatalyst.GetProjectInput{
		Name:      aws.String(id),
		SpaceName: aws.String(spaceName),
	}
	out, err := conn.GetProject(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Name == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

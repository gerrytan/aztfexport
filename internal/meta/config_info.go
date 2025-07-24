package meta

import (
	"io"

	"github.com/Azure/aztfexport/internal/tfaddr"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type ConfigInfos []ConfigInfo

type ConfigInfo struct {
	ImportItem

	dependencies Dependencies

	hcl *hclwrite.File
}

func (cfg ConfigInfo) DumpHCL(w io.Writer) (int, error) {
	out := hclwrite.Format(cfg.hcl.Bytes())
	return w.Write(out)
}

type Dependencies struct {
	// Dependencies inferred by scanning for resource id values, will be applied by substituting with TF address
	// Key is TFResourceId
	refDeps map[string]Dependency

	// Similar to refDeps, but due to multiple Azure resources can map to a same TF resource id, we can't decide which Azure resource
	// is depended on. Hence these will end up as comments inside "depends_on" block for the user to manually resolve.
	// The key is TFResourceId.
	ambiguousRefDeps map[string][]Dependency

	// Inferred by checking if the resource has a common attribute with the same value as its parent. For example if
	// azurerm_virtual_network.res-1.resource_group_name is the same value as its parent azurerm_resource_group.res-0.name,
	// it will be replaced with the TF address.
	// The key is TFResourceId of the child resource.
	commonAttrDeps map[string]CommonAttrDep

	// Dependencies inferred via Azure resource id parent lookup, and will be applied in the "depends_on" block.
	// Dependency that's already satisfied via refDeps should not be included here.
	// TODO: rename this to explicitDeps to avoid confusion with commonAttrDeps?
	parentChildDeps map[Dependency]bool
}

type Dependency struct {
	TFResourceId    string
	AzureResourceId string
	TFAddr          tfaddr.TFAddr
}

type CommonAttrDep struct {
	// AzureResourceId of the parent resource, used to mark that a dependency relation is established
	AzureResourceId string

	// TFAddr of the parent resource, for example: azurerm_resource_group.res-0
	TFAddr tfaddr.TFAddr

	// The attribute name mapping between child -> parent. For example: "resource_group_name" -> "name"
	// This will be used by hcl_edit.go to replace "resource_group_name" = "my-rg" with "resource_group_name" = azurerm_resource_group.res-0.name
	AttrMap map[string]string
}

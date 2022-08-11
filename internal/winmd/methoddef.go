package winmd

import (
	"github.com/tdakkota/win32metadata/md"
	"github.com/tdakkota/win32metadata/types"
)

// GetMethodOverloadName finds and returns the overload attribute for the given method
func GetMethodOverloadName(ctx *types.Context, methodDef *types.MethodDef) string {
	cAttrTable := ctx.Table(md.CustomAttribute)
	for i := uint32(0); i < cAttrTable.RowCount(); i++ {
		var cAttr types.CustomAttribute
		if err := cAttr.FromRow(cAttrTable.Row(i)); err != nil {
			continue
		}

		// - Parent: The owner of the Attribute must be the given func
		if cAttrParentTable, _ := cAttr.Parent.Table(); cAttrParentTable != md.MethodDef {
			continue
		}

		var parentMethodDef types.MethodDef
		row, ok := cAttr.Parent.Row(ctx)
		if !ok {
			continue
		}
		if err := parentMethodDef.FromRow(row); err != nil {
			continue
		}

		// does the blob belong to the method we're looking for?
		if parentMethodDef.Name != methodDef.Name || string(parentMethodDef.Signature) != string(methodDef.Signature) {
			continue
		}

		// - Type: the attribute type must be the given type
		// the cAttr.Type table can be either a MemberRef or a MethodRef.
		// Since we are looking for a type, we will only consider the MemberRef.
		if cAttrTypeTable, _ := cAttr.Type.Table(); cAttrTypeTable != md.MemberRef {
			continue
		}

		var attrTypeMemberRef types.MemberRef
		row, ok = cAttr.Type.Row(ctx)
		if !ok {
			continue
		}
		if err := attrTypeMemberRef.FromRow(row); err != nil {
			continue
		}

		// we need to check the MemberRef Class
		// the value can belong to several tables, but we are only going to check for TypeRef
		if classTable, _ := attrTypeMemberRef.Class.Table(); classTable != md.TypeRef {
			continue
		}

		var attrTypeRef types.TypeRef
		row, ok = attrTypeMemberRef.Class.Row(ctx)
		if !ok {
			continue
		}
		if err := attrTypeRef.FromRow(row); err != nil {
			continue
		}

		if attrTypeRef.TypeNamespace+"."+attrTypeRef.TypeName == AttributeTypeOverloadAttribute {
			// Metadata values start with 0x01 0x00 and ends with 0x00 0x00
			mdVal := cAttr.Value[2 : len(cAttr.Value)-2]
			// the next value is the length of the string
			mdVal = mdVal[1:]
			return string(mdVal)
		}
	}
	return methodDef.Name
}

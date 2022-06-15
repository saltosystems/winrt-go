package metadata

import (
	"fmt"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tdakkota/win32metadata/md"
	"github.com/tdakkota/win32metadata/types"
)

// HasContext is a helper struct that holds the original context of a metadata element.
type HasContext struct {
	originalCtx *types.Context
}

// Ctx return the original context of the element.
func (hctx *HasContext) Ctx() *types.Context {
	return hctx.originalCtx
}

// TypeDef is a helper struct that wraps types.TypeDef and stores the original context
// of the typeDef.
type TypeDef struct {
	types.TypeDef
	HasContext

	logger log.Logger
}

// QualifiedID holds the namespace and the name of a qualified element. This may be a type, a static function or a field
type QualifiedID struct {
	Namespace string
	Name      string
}

// GetValueForEnumField returns the value of the requested enum field.
func (typeDef *TypeDef) GetValueForEnumField(fieldIndex uint32) (string, error) {
	// For each Enum value definition, there is a corresponding row in the Constant table to store the integer value for the enum value.
	tableConstants := typeDef.Ctx().Table(md.Constant)
	for i := uint32(0); i < tableConstants.RowCount(); i++ {
		var constant types.Constant
		if err := constant.FromRow(tableConstants.Row(i)); err != nil {
			return "", err
		}

		if t, _ := constant.Parent.Table(); t != md.Field {
			continue
		}

		// does the blob belong to the field we're looking for?
		// The parent is an index into the field table that holds the associated enum value record
		if constant.Parent.TableIndex() != fieldIndex {
			continue
		}

		// The value is a blob that we need to read as little endian
		var blobIndex uint32
		for i, b := range constant.Value {
			blobIndex += uint32(b) << (i * 8)
		}
		return strconv.Itoa(int(blobIndex)), nil
	}

	return "", fmt.Errorf("no value found for field %d", fieldIndex)
}

// GetTypeDefAttributeWithType returns the value of the given attribute type and fails if not found.
func (typeDef *TypeDef) GetTypeDefAttributeWithType(lookupAttrTypeClass string) ([]byte, error) {
	result := typeDef.GetTypeDefAttributesWithType(lookupAttrTypeClass)
	if len(result) == 0 {
		return nil, fmt.Errorf("type %s has no custom attribute %s", typeDef.TypeNamespace+"."+typeDef.TypeName, lookupAttrTypeClass)
	} else if len(result) > 1 {
		_ = level.Warn(typeDef.logger).Log(
			"msg", "type has multiple custom attributes, returning the first one",
			"type", typeDef.TypeNamespace+"."+typeDef.TypeName,
			"attr", lookupAttrTypeClass,
		)
	}

	return result[0], nil
}

// GetTypeDefAttributesWithType returns the values of all the attributes that match the given type.
func (typeDef *TypeDef) GetTypeDefAttributesWithType(lookupAttrTypeClass string) [][]byte {
	result := make([][]byte, 0)
	cAttrTable := typeDef.Ctx().Table(md.CustomAttribute)
	for i := uint32(0); i < cAttrTable.RowCount(); i++ {
		var cAttr types.CustomAttribute
		if err := cAttr.FromRow(cAttrTable.Row(i)); err != nil {
			continue
		}

		// - Parent: The owner of the Attribute must be the given typeDef
		if cAttrParentTable, _ := cAttr.Parent.Table(); cAttrParentTable != md.TypeDef {
			continue
		}

		var parentTypeDef TypeDef
		row, ok := cAttr.Parent.Row(typeDef.Ctx())
		if !ok {
			continue
		}
		if err := parentTypeDef.FromRow(row); err != nil {
			continue
		}

		// does the blob belong to the type we're looking for?
		if parentTypeDef.TypeNamespace != typeDef.TypeNamespace || parentTypeDef.TypeName != typeDef.TypeName {
			continue
		}

		// - Type: the attribute type must be the given type
		// the cAttr.Type table can be either a MemberRef or a MethodRef.
		// Since we are looking for a type, we will only consider the MemberRef.
		if cAttrTypeTable, _ := cAttr.Type.Table(); cAttrTypeTable != md.MemberRef {
			continue
		}

		var attrTypeMemberRef types.MemberRef
		row, ok = cAttr.Type.Row(typeDef.Ctx())
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
		row, ok = attrTypeMemberRef.Class.Row(typeDef.Ctx())
		if !ok {
			continue
		}
		if err := attrTypeRef.FromRow(row); err != nil {
			continue
		}

		if attrTypeRef.TypeNamespace+"."+attrTypeRef.TypeName == lookupAttrTypeClass {
			result = append(result, cAttr.Value)
		}
	}

	return result
}

// GetImplementedInterfaces returns the interfaces implemented by the type.
func (typeDef *TypeDef) GetImplementedInterfaces() ([]QualifiedID, error) {
	interfaces := make([]QualifiedID, 0)

	tableInterfaceImpl := typeDef.Ctx().Table(md.InterfaceImpl)
	for i := uint32(0); i < tableInterfaceImpl.RowCount(); i++ {
		var interfaceImpl types.InterfaceImpl
		if err := interfaceImpl.FromRow(tableInterfaceImpl.Row(i)); err != nil {
			return nil, err
		}

		classTd, err := interfaceImpl.ResolveClass(typeDef.Ctx())
		if err != nil {
			return nil, err
		}

		if classTd.TypeNamespace+"."+classTd.TypeName != typeDef.TypeNamespace+"."+typeDef.TypeName {
			// not the class we are looking for
			continue
		}

		if t, ok := interfaceImpl.Interface.Table(); ok && t == md.TypeSpec {
			// ignore type spec rows
			continue
		}

		ifaceNS, ifaceName, err := typeDef.Ctx().ResolveTypeDefOrRefName(interfaceImpl.Interface)
		if err != nil {
			return nil, err
		}

		interfaces = append(interfaces, QualifiedID{Namespace: ifaceNS, Name: ifaceName})
	}

	return interfaces, nil
}

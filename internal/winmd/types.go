package winmd

import (
	"fmt"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tdakkota/win32metadata/md"
	"github.com/tdakkota/win32metadata/types"
)

// Custom Attributes
const (
	AttributeTypeGUID                 = "Windows.Foundation.Metadata.GuidAttribute"
	AttributeTypeExclusiveTo          = "Windows.Foundation.Metadata.ExclusiveToAttribute"
	AttributeTypeStaticAttribute      = "Windows.Foundation.Metadata.StaticAttribute"
	AttributeTypeActivatableAttribute = "Windows.Foundation.Metadata.ActivatableAttribute"
	AttributeTypeDefaultAttribute     = "Windows.Foundation.Metadata.DefaultAttribute"
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

// Extends returns true if the type extends the given class
func (typeDef *TypeDef) Extends(class string) (bool, error) {
	ns, name, err := typeDef.Ctx().ResolveTypeDefOrRefName(typeDef.TypeDef.Extends)
	if err != nil {
		return false, err
	}
	return ns+"."+name == class, nil
}

// GetGenericParams returns the generic parameters of the type.
func (typeDef *TypeDef) GetGenericParams() ([]*types.GenericParam, error) {
	params := make([]*types.GenericParam, 0)
	tableGenericParam := typeDef.Ctx().Table(md.GenericParam)
	for i := uint32(0); i < tableGenericParam.RowCount(); i++ {
		var genericParam types.GenericParam
		if err := genericParam.FromRow(tableGenericParam.Row(i)); err != nil {
			continue
		}

		// - Owner: The owner of the Attribute must be the given typeDef
		if genericParamOwnerTable, _ := genericParam.Owner.Table(); genericParamOwnerTable != md.TypeDef {
			continue
		}

		var ownerTypeDef types.TypeDef
		row, ok := genericParam.Owner.Row(typeDef.Ctx())
		if !ok {
			continue
		}
		if err := ownerTypeDef.FromRow(row); err != nil {
			continue
		}

		// does the blob belong to the type we're looking for?
		if ownerTypeDef.TypeNamespace != typeDef.TypeNamespace || ownerTypeDef.TypeName != typeDef.TypeName {
			continue
		}

		params = append(params, &genericParam)
	}
	if len(params) == 0 {
		return nil, fmt.Errorf("could not find generic params for type %s.%s", typeDef.TypeNamespace, typeDef.TypeName)
	}

	return params, nil
}

// IsInterface returns true if the type is an interface
func (typeDef *TypeDef) IsInterface() bool {
	return typeDef.Flags.Interface()
}

// IsEnum returns true if the type is an enum
func (typeDef *TypeDef) IsEnum() bool {
	ok, err := typeDef.Extends("System.Enum")
	if err != nil {
		_ = level.Error(typeDef.logger).Log("msg", "error resolving type extends, all classes should extend at least System.Object", "err", err)
		return false
	}
	return ok
}

// IsDelegate returns true if the type is a delegate
func (typeDef *TypeDef) IsDelegate() bool {
	if !(typeDef.Flags.Public() && typeDef.Flags.Sealed()) {
		return false
	}

	ok, err := typeDef.Extends("System.MulticastDelegate")
	if err != nil {
		_ = level.Error(typeDef.logger).Log("msg", "error resolving type extends, all classes should extend at least System.Object", "err", err)
		return false
	}

	return ok
}

// IsStruct returns true if the type is a struct
func (typeDef *TypeDef) IsStruct() bool {
	ok, err := typeDef.Extends("System.ValueType")
	if err != nil {
		_ = level.Error(typeDef.logger).Log("msg", "error resolving type extends, all classes should extend at least System.Object", "err", err)
		return false
	}
	return ok
}

// IsRuntimeClass returns true if the type is a runtime class
func (typeDef *TypeDef) IsRuntimeClass() bool {
	// Flags: all runtime classes must carry the public, auto layout, class, and tdWindowsRuntime flags.
	return typeDef.Flags.Public() && typeDef.Flags.AutoLayout() && typeDef.Flags.Class() && typeDef.Flags&0x4000 != 0
}

// GUID returns the GUID of the type.
func (typeDef *TypeDef) GUID() (string, error) {
	blob, err := typeDef.GetTypeDefAttributeWithType(AttributeTypeGUID)
	if err != nil {
		return "", err
	}
	return guidBlobToString(blob)
}

// guidBlobToString converts an array into the textual representation of a GUID
func guidBlobToString(b types.Blob) (string, error) {
	// the guid is a blob of 20 bytes
	if len(b) != 20 {
		return "", fmt.Errorf("invalid GUID blob length: %d", len(b))
	}

	// that starts with 0100
	if b[0] != 0x01 || b[1] != 0x00 {
		return "", fmt.Errorf("invalid GUID blob header, expected '0x01 0x00' but found '0x%02x 0x%02x'", b[0], b[1])
	}

	// and ends with 0000
	if b[18] != 0x00 || b[19] != 0x00 {
		return "", fmt.Errorf("invalid GUID blob footer, expected '0x00 0x00' but found '0x%02x 0x%02x'", b[18], b[19])
	}

	guid := b[2 : len(b)-2]
	// the string version has 5 parts separated by '-'
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x",
		// The first 3 are encoded as little endian
		uint32(guid[0])|uint32(guid[1])<<8|uint32(guid[2])<<16|uint32(guid[3])<<24,
		uint16(guid[4])|uint16(guid[5])<<8,
		uint16(guid[6])|uint16(guid[7])<<8,
		//the rest is not
		uint16(guid[8])<<8|uint16(guid[9]),
		uint16(guid[10])<<8|uint16(guid[11]),
		uint32(guid[12])<<24|uint32(guid[13])<<16|uint32(guid[14])<<8|uint32(guid[15])), nil
}

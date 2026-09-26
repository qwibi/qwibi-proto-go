package qwibiv1

import (
	"testing"

	validatepb "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	uuidPattern         = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	optionalUUIDPattern = "^$|" + uuidPattern
)

func TestLF04aLayerWireContract(t *testing.T) {
	layer := File_qwibi_v1_layer_proto.Messages().ByName("GeoLayer")
	if layer == nil {
		t.Fatal("GeoLayer descriptor is missing")
	}

	assertUUIDField(t, layer, "organization_id", 17, uuidPattern, 36, false)
	assertUUIDField(t, layer, "created_by_id", 18, uuidPattern, 36, false)
	assertFieldAbsentAndReserved(t, layer, "app_id", 12)

	createLayer := File_qwibi_v1_layer_proto.Messages().ByName("CreateLayerRequest")
	if createLayer == nil {
		t.Fatal("CreateLayerRequest descriptor is missing")
	}
	assertFieldAbsentAndReserved(t, createLayer, "app_id", 8)
	assertUUIDField(t, createLayer, "organization_id", 12, uuidPattern, 36, false)
	assertUUIDField(t, createLayer, "on_behalf_of_account_id", 11, optionalUUIDPattern, 0, true)
	if field := createLayer.Fields().ByName("created_by_id"); field != nil {
		t.Error("CreateLayerRequest.created_by_id unexpectedly exists")
	}

	transfer := File_qwibi_v1_layer_proto.Messages().ByName("TransferLayerOwnershipRequest")
	if transfer == nil {
		t.Fatal("TransferLayerOwnershipRequest descriptor is missing")
	}
	if transfer.Fields().Len() != 2 {
		t.Fatalf("TransferLayerOwnershipRequest fields = %d, want exactly 2", transfer.Fields().Len())
	}
	assertUUIDField(t, transfer, "layer_id", 1, uuidPattern, 36, false)
	assertUUIDField(t, transfer, "new_owner_account_id", 2, uuidPattern, 36, false)

	transferResponse := File_qwibi_v1_layer_proto.Messages().ByName("TransferLayerOwnershipResponse")
	if transferResponse == nil {
		t.Fatal("TransferLayerOwnershipResponse descriptor is missing")
	}
	if transferResponse.Fields().Len() != 1 {
		t.Fatalf("TransferLayerOwnershipResponse fields = %d, want exactly 1", transferResponse.Fields().Len())
	}
	layerField := transferResponse.Fields().ByName("layer")
	if layerField == nil ||
		layerField.Number() != 1 ||
		layerField.Kind() != protoreflect.MessageKind ||
		layerField.Message().FullName() != "qwibi.v1.GeoLayer" {
		t.Fatalf("TransferLayerOwnershipResponse.layer descriptor = %v, want qwibi.v1.GeoLayer field 1", layerField)
	}

	service := File_qwibi_v1_service_proto.Services().ByName("QwibiService")
	if service == nil {
		t.Fatal("QwibiService descriptor is missing")
	}
	method := service.Methods().ByName("TransferLayerOwnership")
	if method == nil {
		t.Fatal("TransferLayerOwnership RPC descriptor is missing")
	}
	if method.IsStreamingClient() || method.IsStreamingServer() {
		t.Fatal("TransferLayerOwnership must be unary")
	}
	if method.Input().FullName() != "qwibi.v1.TransferLayerOwnershipRequest" ||
		method.Output().FullName() != "qwibi.v1.TransferLayerOwnershipResponse" {
		t.Fatalf("TransferLayerOwnership types = %s -> %s", method.Input().FullName(), method.Output().FullName())
	}
	level, ok := proto.GetExtension(method.Options(), E_Auth).(AuthLevel)
	if !ok || level != AuthLevel_AUTH_LEVEL_REQUIRED {
		t.Fatalf("TransferLayerOwnership auth = %v, present=%t; want AUTH_LEVEL_REQUIRED", level, ok)
	}
}

func assertFieldAbsentAndReserved(
	t *testing.T,
	message protoreflect.MessageDescriptor,
	name protoreflect.Name,
	number protoreflect.FieldNumber,
) {
	t.Helper()

	if field := message.Fields().ByName(name); field != nil {
		t.Errorf("%s.%s unexpectedly exists", message.Name(), name)
	}
	if field := message.Fields().ByNumber(number); field != nil {
		t.Errorf("%s field number %d unexpectedly used by %s", message.Name(), number, field.Name())
	}
	if !message.ReservedRanges().Has(number) {
		t.Errorf("%s field number %d is not reserved", message.Name(), number)
	}
	if !message.ReservedNames().Has(name) {
		t.Errorf("%s field name %s is not reserved", message.Name(), name)
	}
}

func assertUUIDField(
	t *testing.T,
	message protoreflect.MessageDescriptor,
	name protoreflect.Name,
	number protoreflect.FieldNumber,
	pattern string,
	length uint64,
	deprecated bool,
) {
	t.Helper()

	field := message.Fields().ByName(name)
	if field == nil {
		t.Errorf("%s.%s descriptor is missing", message.Name(), name)
		return
	}
	if field.Number() != number {
		t.Errorf("%s.%s number = %d, want %d", message.Name(), name, field.Number(), number)
	}
	if field.Kind() != protoreflect.StringKind {
		t.Errorf("%s.%s kind = %s, want string", message.Name(), name, field.Kind())
	}

	options, ok := field.Options().(*descriptorpb.FieldOptions)
	if !ok {
		t.Fatalf("%s.%s options have type %T", message.Name(), name, field.Options())
	}
	if options.GetDeprecated() != deprecated {
		t.Errorf("%s.%s deprecated = %t, want %t", message.Name(), name, options.GetDeprecated(), deprecated)
	}
	if !proto.HasExtension(options, validatepb.E_Field) {
		t.Fatalf("%s.%s has no protovalidate field rules", message.Name(), name)
	}
	rules, ok := proto.GetExtension(options, validatepb.E_Field).(*validatepb.FieldRules)
	if !ok {
		t.Fatalf("%s.%s protovalidate rules have unexpected type", message.Name(), name)
	}
	stringRules := rules.GetString()
	if stringRules == nil {
		t.Fatalf("%s.%s has no string rules", message.Name(), name)
	}
	if stringRules.GetPattern() != pattern {
		t.Errorf("%s.%s pattern = %q, want %q", message.Name(), name, stringRules.GetPattern(), pattern)
	}
	if stringRules.GetLen() != length {
		t.Errorf("%s.%s len = %d, want %d", message.Name(), name, stringRules.GetLen(), length)
	}
}

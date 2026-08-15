// This is a generated file - do not edit.
//
// Generated from echo/v1/echo.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use echoRequestDescriptor instead')
const EchoRequest$json = {
  '1': 'EchoRequest',
  '2': [
    {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
    {'1': 'repeat', '3': 2, '4': 1, '5': 13, '10': 'repeat'},
  ],
};

/// Descriptor for `EchoRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List echoRequestDescriptor = $convert.base64Decode(
    'CgtFY2hvUmVxdWVzdBIYCgdtZXNzYWdlGAEgASgJUgdtZXNzYWdlEhYKBnJlcGVhdBgCIAEoDV'
    'IGcmVwZWF0');

@$core.Deprecated('Use echoResponseDescriptor instead')
const EchoResponse$json = {
  '1': 'EchoResponse',
  '2': [
    {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
    {'1': 'unix_millis', '3': 2, '4': 1, '5': 3, '10': 'unixMillis'},
  ],
};

/// Descriptor for `EchoResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List echoResponseDescriptor = $convert.base64Decode(
    'CgxFY2hvUmVzcG9uc2USGAoHbWVzc2FnZRgBIAEoCVIHbWVzc2FnZRIfCgt1bml4X21pbGxpcx'
    'gCIAEoA1IKdW5peE1pbGxpcw==');

@$core.Deprecated('Use reverseRequestDescriptor instead')
const ReverseRequest$json = {
  '1': 'ReverseRequest',
  '2': [
    {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
  ],
};

/// Descriptor for `ReverseRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reverseRequestDescriptor = $convert
    .base64Decode('Cg5SZXZlcnNlUmVxdWVzdBIYCgdtZXNzYWdlGAEgASgJUgdtZXNzYWdl');

@$core.Deprecated('Use reverseResponseDescriptor instead')
const ReverseResponse$json = {
  '1': 'ReverseResponse',
  '2': [
    {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
    {'1': 'unix_millis', '3': 2, '4': 1, '5': 3, '10': 'unixMillis'},
  ],
};

/// Descriptor for `ReverseResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reverseResponseDescriptor = $convert.base64Decode(
    'Cg9SZXZlcnNlUmVzcG9uc2USGAoHbWVzc2FnZRgBIAEoCVIHbWVzc2FnZRIfCgt1bml4X21pbG'
    'xpcxgCIAEoA1IKdW5peE1pbGxpcw==');

const $core.Map<$core.String, $core.dynamic> EchoServiceBase$json = {
  '1': 'EchoService',
  '2': [
    {'1': 'Echo', '2': '.echo.v1.EchoRequest', '3': '.echo.v1.EchoResponse'},
    {
      '1': 'Reverse',
      '2': '.echo.v1.ReverseRequest',
      '3': '.echo.v1.ReverseResponse'
    },
  ],
};

@$core.Deprecated('Use echoServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    EchoServiceBase$messageJson = {
  '.echo.v1.EchoRequest': EchoRequest$json,
  '.echo.v1.EchoResponse': EchoResponse$json,
  '.echo.v1.ReverseRequest': ReverseRequest$json,
  '.echo.v1.ReverseResponse': ReverseResponse$json,
};

/// Descriptor for `EchoService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List echoServiceDescriptor = $convert.base64Decode(
    'CgtFY2hvU2VydmljZRIzCgRFY2hvEhQuZWNoby52MS5FY2hvUmVxdWVzdBoVLmVjaG8udjEuRW'
    'Nob1Jlc3BvbnNlEjwKB1JldmVyc2USFy5lY2hvLnYxLlJldmVyc2VSZXF1ZXN0GhguZWNoby52'
    'MS5SZXZlcnNlUmVzcG9uc2U=');

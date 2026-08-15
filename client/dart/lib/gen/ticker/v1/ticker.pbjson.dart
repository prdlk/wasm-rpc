// This is a generated file - do not edit.
//
// Generated from ticker/v1/ticker.proto.

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

@$core.Deprecated('Use subscribeRequestDescriptor instead')
const SubscribeRequest$json = {
  '1': 'SubscribeRequest',
  '2': [
    {'1': 'topic', '3': 1, '4': 1, '5': 9, '10': 'topic'},
    {'1': 'count', '3': 2, '4': 1, '5': 13, '10': 'count'},
    {'1': 'interval_ms', '3': 3, '4': 1, '5': 13, '10': 'intervalMs'},
  ],
};

/// Descriptor for `SubscribeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List subscribeRequestDescriptor = $convert.base64Decode(
    'ChBTdWJzY3JpYmVSZXF1ZXN0EhQKBXRvcGljGAEgASgJUgV0b3BpYxIUCgVjb3VudBgCIAEoDV'
    'IFY291bnQSHwoLaW50ZXJ2YWxfbXMYAyABKA1SCmludGVydmFsTXM=');

@$core.Deprecated('Use subscribeResponseDescriptor instead')
const SubscribeResponse$json = {
  '1': 'SubscribeResponse',
  '2': [
    {'1': 'topic', '3': 1, '4': 1, '5': 9, '10': 'topic'},
    {'1': 'seq', '3': 2, '4': 1, '5': 13, '10': 'seq'},
    {'1': 'unix_millis', '3': 3, '4': 1, '5': 3, '10': 'unixMillis'},
  ],
};

/// Descriptor for `SubscribeResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List subscribeResponseDescriptor = $convert.base64Decode(
    'ChFTdWJzY3JpYmVSZXNwb25zZRIUCgV0b3BpYxgBIAEoCVIFdG9waWMSEAoDc2VxGAIgASgNUg'
    'NzZXESHwoLdW5peF9taWxsaXMYAyABKANSCnVuaXhNaWxsaXM=');

const $core.Map<$core.String, $core.dynamic> TickerServiceBase$json = {
  '1': 'TickerService',
  '2': [
    {
      '1': 'Subscribe',
      '2': '.ticker.v1.SubscribeRequest',
      '3': '.ticker.v1.SubscribeResponse',
      '6': true
    },
  ],
};

@$core.Deprecated('Use tickerServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    TickerServiceBase$messageJson = {
  '.ticker.v1.SubscribeRequest': SubscribeRequest$json,
  '.ticker.v1.SubscribeResponse': SubscribeResponse$json,
};

/// Descriptor for `TickerService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List tickerServiceDescriptor = $convert.base64Decode(
    'Cg1UaWNrZXJTZXJ2aWNlEkgKCVN1YnNjcmliZRIbLnRpY2tlci52MS5TdWJzY3JpYmVSZXF1ZX'
    'N0GhwudGlja2VyLnYxLlN1YnNjcmliZVJlc3BvbnNlMAE=');

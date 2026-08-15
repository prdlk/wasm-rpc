// This is a generated file - do not edit.
//
// Generated from echo/v1/echo.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class EchoRequest extends $pb.GeneratedMessage {
  factory EchoRequest({
    $core.String? message,
    $core.int? repeat,
  }) {
    final result = create();
    if (message != null) result.message = message;
    if (repeat != null) result.repeat = repeat;
    return result;
  }

  EchoRequest._();

  factory EchoRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory EchoRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EchoRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'echo.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'message')
    ..aI(2, _omitFieldNames ? '' : 'repeat', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EchoRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EchoRequest copyWith(void Function(EchoRequest) updates) =>
      super.copyWith((message) => updates(message as EchoRequest))
          as EchoRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EchoRequest create() => EchoRequest._();
  @$core.override
  EchoRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static EchoRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EchoRequest>(create);
  static EchoRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => $_clearField(1);

  /// If > 1, the message is repeated N times.
  @$pb.TagNumber(2)
  $core.int get repeat => $_getIZ(1);
  @$pb.TagNumber(2)
  set repeat($core.int value) => $_setUnsignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRepeat() => $_has(1);
  @$pb.TagNumber(2)
  void clearRepeat() => $_clearField(2);
}

class EchoResponse extends $pb.GeneratedMessage {
  factory EchoResponse({
    $core.String? message,
    $fixnum.Int64? unixMillis,
  }) {
    final result = create();
    if (message != null) result.message = message;
    if (unixMillis != null) result.unixMillis = unixMillis;
    return result;
  }

  EchoResponse._();

  factory EchoResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory EchoResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EchoResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'echo.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'message')
    ..aInt64(2, _omitFieldNames ? '' : 'unixMillis')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EchoResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EchoResponse copyWith(void Function(EchoResponse) updates) =>
      super.copyWith((message) => updates(message as EchoResponse))
          as EchoResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EchoResponse create() => EchoResponse._();
  @$core.override
  EchoResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static EchoResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EchoResponse>(create);
  static EchoResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => $_clearField(1);

  /// Server-side timestamp (unix millis) proving the response was
  /// produced inside the Go handler.
  @$pb.TagNumber(2)
  $fixnum.Int64 get unixMillis => $_getI64(1);
  @$pb.TagNumber(2)
  set unixMillis($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUnixMillis() => $_has(1);
  @$pb.TagNumber(2)
  void clearUnixMillis() => $_clearField(2);
}

class ReverseRequest extends $pb.GeneratedMessage {
  factory ReverseRequest({
    $core.String? message,
  }) {
    final result = create();
    if (message != null) result.message = message;
    return result;
  }

  ReverseRequest._();

  factory ReverseRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ReverseRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReverseRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'echo.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReverseRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReverseRequest copyWith(void Function(ReverseRequest) updates) =>
      super.copyWith((message) => updates(message as ReverseRequest))
          as ReverseRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ReverseRequest create() => ReverseRequest._();
  @$core.override
  ReverseRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ReverseRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReverseRequest>(create);
  static ReverseRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => $_clearField(1);
}

class ReverseResponse extends $pb.GeneratedMessage {
  factory ReverseResponse({
    $core.String? message,
    $fixnum.Int64? unixMillis,
  }) {
    final result = create();
    if (message != null) result.message = message;
    if (unixMillis != null) result.unixMillis = unixMillis;
    return result;
  }

  ReverseResponse._();

  factory ReverseResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ReverseResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReverseResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'echo.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'message')
    ..aInt64(2, _omitFieldNames ? '' : 'unixMillis')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReverseResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReverseResponse copyWith(void Function(ReverseResponse) updates) =>
      super.copyWith((message) => updates(message as ReverseResponse))
          as ReverseResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ReverseResponse create() => ReverseResponse._();
  @$core.override
  ReverseResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ReverseResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReverseResponse>(create);
  static ReverseResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get unixMillis => $_getI64(1);
  @$pb.TagNumber(2)
  set unixMillis($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUnixMillis() => $_has(1);
  @$pb.TagNumber(2)
  void clearUnixMillis() => $_clearField(2);
}

/// EchoService is a generic service used to validate the full
/// client -> Wasm -> Go -> Wasm -> client pipeline.
class EchoServiceApi {
  final $pb.RpcClient _client;

  EchoServiceApi(this._client);

  /// Echo returns the message, optionally repeated.
  $async.Future<EchoResponse> echo(
          $pb.ClientContext? ctx, EchoRequest request) =>
      _client.invoke<EchoResponse>(
          ctx, 'EchoService', 'Echo', request, EchoResponse());

  /// Reverse returns the message reversed (rune-safe).
  $async.Future<ReverseResponse> reverse(
          $pb.ClientContext? ctx, ReverseRequest request) =>
      _client.invoke<ReverseResponse>(
          ctx, 'EchoService', 'Reverse', request, ReverseResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');

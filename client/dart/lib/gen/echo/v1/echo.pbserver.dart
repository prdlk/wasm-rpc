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

import 'package:protobuf/protobuf.dart' as $pb;

import 'echo.pb.dart' as $0;
import 'echo.pbjson.dart';

export 'echo.pb.dart';

abstract class EchoServiceBase extends $pb.GeneratedService {
  $async.Future<$0.EchoResponse> echo(
      $pb.ServerContext ctx, $0.EchoRequest request);
  $async.Future<$0.ReverseResponse> reverse(
      $pb.ServerContext ctx, $0.ReverseRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'Echo':
        return $0.EchoRequest();
      case 'Reverse':
        return $0.ReverseRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'Echo':
        return echo(ctx, request as $0.EchoRequest);
      case 'Reverse':
        return reverse(ctx, request as $0.ReverseRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => EchoServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => EchoServiceBase$messageJson;
}

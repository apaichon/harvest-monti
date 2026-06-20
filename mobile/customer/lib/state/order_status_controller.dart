import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

import '../api/api_client.dart';
import '../api/order_api.dart';
import '../models/order.dart';

/// Order status surface for `/api/v1/public/orders/{order_number}/stream`.
///
/// Behaviour per DES-0007 §7 and REQ-0010 AC-9:
/// - Subscribe to WS first; on any close/error fall back to 10s polling.
/// - Reconnect path: if the WS closes mid-flow we keep showing the last
///   known status and just start polling — no error UI per TEST-0011 TC-10.
@immutable
class OrderStatusState {
  const OrderStatusState({
    required this.status,
    this.connection = ConnectionMode.connecting,
  });

  final OrderStatus status;
  final ConnectionMode connection;

  OrderStatusState copyWith({OrderStatus? status, ConnectionMode? connection}) =>
      OrderStatusState(
        status: status ?? this.status,
        connection: connection ?? this.connection,
      );
}

enum ConnectionMode { connecting, websocket, polling, offline }

typedef WsConnector = WebSocketChannel Function(Uri uri);

class OrderStatusController extends StateNotifier<OrderStatusState> {
  OrderStatusController({
    required this.orderNumber,
    required this.orderApi,
    WsConnector? wsConnector,
    Duration pollInterval = const Duration(seconds: 10),
  })  : _wsConnect = wsConnector ?? WebSocketChannel.connect,
        _pollInterval = pollInterval,
        super(
          OrderStatusState(
            status: OrderStatus(
              orderNumber: orderNumber,
              currentStep: OrderStep.received,
            ),
          ),
        );

  final String orderNumber;
  final OrderApi orderApi;
  final WsConnector _wsConnect;
  final Duration _pollInterval;

  StreamSubscription? _wsSub;
  Timer? _pollTimer;
  WebSocketChannel? _channel;

  /// Connect (called by the screen's initState).
  void connect() {
    _openWs();
  }

  void _openWs() {
    final base = ApiClient().baseUrl
        .replaceFirst('https://', 'wss://')
        .replaceFirst('http://', 'ws://');
    final uri = Uri.parse('$base/api/v1/public/orders/$orderNumber/stream');
    try {
      _channel = _wsConnect(uri);
      state = state.copyWith(connection: ConnectionMode.websocket);
      _wsSub = _channel!.stream.listen(
        _onWsFrame,
        onError: (_) => _fallbackToPolling(),
        onDone: _fallbackToPolling,
        cancelOnError: true,
      );
    } catch (_) {
      _fallbackToPolling();
    }
  }

  void _onWsFrame(dynamic raw) {
    try {
      final m = jsonDecode(raw as String) as Map<String, dynamic>;
      final next = OrderStatus.fromJson(m);
      state = state.copyWith(status: next);
    } catch (_) {
      // ignore malformed frame; do not show error UI per TEST-0011 TC-10.
    }
  }

  void _fallbackToPolling() {
    _wsSub?.cancel();
    _wsSub = null;
    _channel = null;
    if (state.connection == ConnectionMode.polling) return;
    state = state.copyWith(connection: ConnectionMode.polling);
    _pollTimer?.cancel();
    _pollTimer = Timer.periodic(_pollInterval, (_) => _pollOnce());
    _pollOnce();
  }

  Future<void> _pollOnce() async {
    try {
      final s = await orderApi.fetchStatus(orderNumber);
      state = state.copyWith(status: s);
    } catch (_) {
      // Stay in polling; flip to offline if needed.
      state = state.copyWith(connection: ConnectionMode.offline);
    }
  }

  /// Test-only injection — push a status update as if it came from WS.
  @visibleForTesting
  void pushStatus(OrderStatus s) => state = state.copyWith(status: s);

  /// Test-only — force the polling fallback path.
  @visibleForTesting
  void simulateWsDrop() => _fallbackToPolling();

  @override
  void dispose() {
    _wsSub?.cancel();
    _pollTimer?.cancel();
    _channel?.sink.close();
    super.dispose();
  }
}

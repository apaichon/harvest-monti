import 'dart:convert';

import '../models/order.dart';
import 'api_client.dart';

/// Order REST contract from TASK-0009.
class OrderApi {
  OrderApi(this._client);
  final ApiClient _client;

  /// Submits the cart. Backend may return `payment_method='cash'` for
  /// pay-at-counter; no real settlement happens in MVP (ADR-0005 D2 / REQ-0010).
  Future<String> submit({
    required String sessionId,
    required String tableCode,
    required PaymentMethod method,
    required String idempotencyKey,
  }) async {
    final r = await _client.post(
      '/api/v1/public/orders',
      body: jsonEncode({
        'session_id': sessionId,
        'table_code': tableCode,
        'payment_method': method.apiValue,
        'idempotency_key': idempotencyKey,
      }),
    );
    if (r.statusCode != 200 && r.statusCode != 201) {
      throw Exception('orders failed: ${r.statusCode} ${r.body}');
    }
    final m = jsonDecode(r.body) as Map<String, dynamic>;
    return m['order_number'] as String;
  }

  /// Polling fallback per DES-0007 §7 — used when the WS stream closes.
  Future<OrderStatus> fetchStatus(String orderNumber) async {
    final r = await _client.get('/api/v1/public/orders/$orderNumber');
    if (r.statusCode != 200) {
      throw Exception('orders/$orderNumber failed: ${r.statusCode}');
    }
    return OrderStatus.fromJson(jsonDecode(r.body) as Map<String, dynamic>);
  }
}

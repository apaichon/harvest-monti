import 'dart:convert';

import 'api_client.dart';

/// `POST /api/v1/public/qr/redeem` — binds an opaque JWT QR token to a
/// `{session_id, tenant_id, outlet_id, table_code}` per DES-0007 §7.
class QrApi {
  QrApi(this._client);
  final ApiClient _client;

  Future<QrRedeemResult> redeem(String jwt) async {
    final r = await _client.post(
      '/api/v1/public/qr/redeem',
      body: jsonEncode({'token': jwt}),
    );
    if (r.statusCode != 200) {
      throw Exception('qr/redeem failed: ${r.statusCode} ${r.body}');
    }
    final m = jsonDecode(r.body) as Map<String, dynamic>;
    return QrRedeemResult(
      sessionId: m['session_id'] as String,
      tenantId: m['tenant_id'] as String,
      outletId: m['outlet_id'] as String,
      tableCode: m['table_code'] as String,
    );
  }
}

class QrRedeemResult {
  const QrRedeemResult({
    required this.sessionId,
    required this.tenantId,
    required this.outletId,
    required this.tableCode,
  });
  final String sessionId;
  final String tenantId;
  final String outletId;
  final String tableCode;
}

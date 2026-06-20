import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../api/api_client.dart';
import '../api/menu_api.dart';
import '../api/order_api.dart';
import '../api/qr_api.dart';
import '../models/cart.dart';
import '../storage/language_pref.dart';
import 'cart_controller.dart';
import 'order_status_controller.dart';

/// Locale state. Persisted via [LanguagePref].
final localeProvider = StateNotifierProvider<LocaleController, Locale>(
  (ref) => LocaleController(ref),
);

class LocaleController extends StateNotifier<Locale> {
  LocaleController(this.ref) : super(const Locale('en'));
  final Ref ref;

  Future<void> hydrate() async {
    final pref = await LanguagePref.open();
    final code = pref.read();
    if (code != null && code.isNotEmpty) state = Locale(code);
  }

  Future<void> set(Locale locale) async {
    state = locale;
    final pref = await LanguagePref.open();
    await pref.write(locale.languageCode);
  }
}

/// Currently bound table (from QR redeem). Null until QR scan or manual entry.
final tableCodeProvider = StateProvider<String?>((_) => null);

/// Optional voice/session id from QR redeem; used by voice WS handshake too.
final sessionIdProvider = StateProvider<String?>((_) => null);

// REST clients.
final apiClientProvider = Provider<ApiClient>((_) => ApiClient());
final menuApiProvider = Provider<MenuApi>((ref) => MenuApi(ref.read(apiClientProvider)));
final orderApiProvider = Provider<OrderApi>((ref) => OrderApi(ref.read(apiClientProvider)));
final qrApiProvider = Provider<QrApi>((ref) => QrApi(ref.read(apiClientProvider)));

// Cart.
final cartProvider =
    StateNotifierProvider<CartController, CartState>((ref) => CartController());

// Order status.
final orderStatusProvider = StateNotifierProvider.family<
    OrderStatusController, OrderStatusState, String>(
  (ref, orderNumber) => OrderStatusController(
    orderNumber: orderNumber,
    orderApi: ref.read(orderApiProvider),
  ),
);

/// Re-exported for screen ergonomics.
typedef CartTotalsSnapshot = CartTotals;

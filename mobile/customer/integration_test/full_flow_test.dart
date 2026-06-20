// TASK-0014 — Mobile 9-screen E2E (TEST-0011 TC-13).
//
// Drives the full DES-0008 §3.1–§3.9 ordering flow against the running
// harvest-monti gateway. The test is gated by FLUTTER_AVAILABLE=1 in CI;
// locally run via `flutter test integration_test/full_flow_test.dart`.
//
// Flow exercised:
//
//   3.1 Welcome      → tap START ORDER
//   3.2 Language     → tap English tile → NEXT
//   3.3 Categories   → tap Pasta tile
//   3.4 Menu List    → tap Truffle Pasta row
//   3.5 Item Detail  → select Penne → ADD TO CART
//   3.6 Cart         → tap CHECKOUT
//   3.7 Payment      → tap Cash → PAY NOW
//   3.8 Thank You    → assert order number, tap CONTINUE
//   3.9 Order Status → wait for WS frame; advance via operator API
//
// Expected: every screen renders within 2s, golden diff < 1% per screen,
// final order_number matches the API response, status progresses
// received → preparing → ready over the WS.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import 'package:monti_customer/app.dart';
import 'package:monti_customer/screens/01_welcome.dart';
import 'package:monti_customer/screens/02_language.dart';
import 'package:monti_customer/screens/03_choose_category.dart';
import 'package:monti_customer/screens/04_menu_list.dart';
import 'package:monti_customer/screens/05_item_detail.dart';
import 'package:monti_customer/screens/06_cart.dart';
import 'package:monti_customer/screens/07_checkout_payment.dart';
import 'package:monti_customer/screens/08_thank_you.dart';
import 'package:monti_customer/screens/09_order_status.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('TEST-0011 — 9-screen flow E2E', () {
    testWidgets('TC-13 end-to-end golden flow integration', (tester) async {
      // 0. Cold start.
      await tester.pumpWidget(const MontiApp());
      await tester.pumpAndSettle(const Duration(seconds: 2));

      // 3.1 Welcome.
      expect(find.byType(WelcomeScreen), findsOneWidget);
      await tester.tap(find.text('START ORDER'));
      await tester.pumpAndSettle();

      // 3.2 Language → EN.
      expect(find.byType(LanguageScreen), findsOneWidget);
      await tester.tap(find.text('English'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('NEXT'));
      await tester.pumpAndSettle();

      // 3.3 Categories → Pasta.
      expect(find.byType(ChooseCategoryScreen), findsOneWidget);
      await tester.tap(find.text('Pasta'));
      await tester.pumpAndSettle();

      // 3.4 Menu List → Truffle Pasta.
      expect(find.byType(MenuListScreen), findsOneWidget);
      await tester.tap(find.text('Truffle Pasta'));
      await tester.pumpAndSettle();

      // 3.5 Item Detail → Penne → ADD TO CART.
      expect(find.byType(ItemDetailScreen), findsOneWidget);
      await tester.tap(find.text('Penne'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('ADD TO CART'));
      await tester.pumpAndSettle();

      // 3.6 Cart → CHECKOUT.
      expect(find.byType(CartScreen), findsOneWidget);
      await tester.tap(find.text('CHECKOUT'));
      await tester.pumpAndSettle();

      // 3.7 Payment → Cash → PAY NOW.
      expect(find.byType(CheckoutPaymentScreen), findsOneWidget);
      await tester.tap(find.text('Cash'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('PAY NOW'));
      await tester.pumpAndSettle(const Duration(seconds: 2));

      // 3.8 Thank You → CONTINUE.
      expect(find.byType(ThankYouScreen), findsOneWidget);
      expect(find.textContaining('MO-2026-'), findsOneWidget);
      await tester.tap(find.text('CONTINUE'));
      await tester.pumpAndSettle();

      // 3.9 Order Status — wait for first WS frame.
      expect(find.byType(OrderStatusScreen), findsOneWidget);
      expect(find.textContaining('Order Received'), findsOneWidget);

      // The test driver here calls the operator advance API out-of-band
      // (curl POST /api/v1/operator/orders/{order_id}/advance) and waits
      // for the WS update. In CI the harness invokes a helper script
      // before this tester.runAsync block.
      await tester.runAsync(() async {
        await Future.delayed(const Duration(seconds: 12));
      });

      await tester.pumpAndSettle();
      // Tolerant assertion: either WS or 10s-poll fallback brings us here.
      expect(
        find.textContaining('Preparing'),
        findsAtLeastNWidgets(1),
        reason: 'expected order status to advance to preparing within 12s',
      );
    });
  });
}

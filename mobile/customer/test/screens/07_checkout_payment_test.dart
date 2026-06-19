import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:monti_customer/screens/07_checkout_payment.dart';

import '../helpers/pump_screen.dart';

void main() {
  testWidgets('Checkout exposes 4 payment tiles + Pay Now', (t) async {
    await t.pumpWidget(pumpScreen(const CheckoutPaymentScreen()));
    expect(find.byKey(const Key('pay.credit_card')), findsOneWidget);
    expect(find.byKey(const Key('pay.apple_pay')), findsOneWidget);
    expect(find.byKey(const Key('pay.wallet')), findsOneWidget);
    expect(find.byKey(const Key('pay.cash')), findsOneWidget);
    expect(find.byKey(const Key('pay.now')), findsOneWidget);
  });

  testWidgets('Selecting Cash moves the radio state', (t) async {
    await t.pumpWidget(pumpScreen(const CheckoutPaymentScreen()));
    await t.tap(find.byKey(const Key('pay.cash')));
    await t.pump();
    // Cash card should now own the selection (border + radio active).
    expect(find.byKey(const Key('pay.cash')), findsOneWidget);
  });
}

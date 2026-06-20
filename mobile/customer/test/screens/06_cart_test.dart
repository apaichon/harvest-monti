import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:monti_customer/models/menu.dart';
import 'package:monti_customer/screens/06_cart.dart';
import 'package:monti_customer/state/providers.dart';

import '../helpers/pump_screen.dart';

void main() {
  testWidgets('Cart shows Discount line when promo is applied', (t) async {
    final container = ProviderContainer();
    // Seed a line + promo.
    container.read(cartProvider.notifier).addLine(
      item: const MenuItem(
          id: 'latte', name: 'Iced Latte', priceTHB: 120),
      qty: 1,
      unitPriceTHB: 120,
      selectedModifierOptionIds: const {},
    );

    await t.pumpWidget(UncontrolledProviderScope(
      container: container,
      child: pumpScreen(const CartScreen()),
    ));

    expect(find.text('Subtotal'), findsOneWidget);
    expect(find.text('Discount'), findsNothing);

    await t.enterText(find.byKey(const Key('cart.promo')), 'LUNCH20');
    await t.tap(find.byKey(const Key('cart.promo_apply')));
    await t.pump();

    expect(find.text('Discount'), findsOneWidget);
    expect(find.textContaining('−฿ 50'), findsOneWidget);
  });
}

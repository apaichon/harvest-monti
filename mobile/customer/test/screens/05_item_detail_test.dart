import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:monti_customer/models/menu.dart';
import 'package:monti_customer/screens/05_item_detail.dart';

import '../helpers/pump_screen.dart';

const _margherita = MenuItem(
  id: 'margherita',
  name: 'Margherita Pizza',
  description: 'Tomato, mozzarella, basil.',
  priceTHB: 240,
  modifierGroups: [
    ModifierGroup(
      id: 'side',
      label: 'Add Side',
      selectionKind: ModifierSelectionKind.single,
      required: true,
      options: [
        ModifierOption(id: 'salad', label: 'Salad'),
        ModifierOption(id: 'soup', label: 'Soup'),
        ModifierOption(id: 'bread', label: 'Bread'),
      ],
    ),
  ],
);

void main() {
  testWidgets('Add To Cart is disabled until required modifier is picked',
      (t) async {
    await t.pumpWidget(pumpScreen(const ItemDetailScreen(item: _margherita)));
    final btn0 = t.widget<ElevatedButton>(
        find.byKey(const Key('item.add_to_cart')));
    expect(btn0.onPressed, isNull);

    await t.tap(find.byKey(const Key('item.mod.side.salad')));
    await t.pump();
    final btn1 = t.widget<ElevatedButton>(
        find.byKey(const Key('item.add_to_cart')));
    expect(btn1.onPressed, isNotNull);
  });

  testWidgets('Qty stepper updates running total in CTA label', (t) async {
    await t.pumpWidget(pumpScreen(const ItemDetailScreen(item: _margherita)));
    await t.tap(find.byKey(const Key('item.mod.side.salad')));
    await t.pump();
    await t.tap(find.byKey(const Key('item.qty_plus')));
    await t.pump();
    expect(find.textContaining('฿ 480'), findsOneWidget);
  });
}

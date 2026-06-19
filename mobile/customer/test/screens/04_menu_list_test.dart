import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:monti_customer/screens/04_menu_list.dart';

import '../helpers/pump_screen.dart';

void main() {
  testWidgets('Menu list renders rows and supports search filter', (t) async {
    await t.pumpWidget(pumpScreen(const MenuListScreen(categoryId: 'pizza')));
    await t.pump(const Duration(milliseconds: 50));

    expect(find.byKey(const Key('menu.search')), findsOneWidget);
    expect(find.byKey(const Key('menu.row.margherita')), findsOneWidget);
    expect(find.byKey(const Key('menu.row.pepperoni')), findsOneWidget);

    await t.enterText(find.byKey(const Key('menu.search')), 'marg');
    await t.pump();
    expect(find.byKey(const Key('menu.row.margherita')), findsOneWidget);
    expect(find.byKey(const Key('menu.row.pepperoni')), findsNothing);
  });
}

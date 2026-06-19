import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:monti_customer/screens/03_choose_category.dart';

import '../helpers/pump_screen.dart';

void main() {
  testWidgets('Categories renders 2x3 grid from fallback when API fails',
      (t) async {
    await t.pumpWidget(pumpScreen(const ChooseCategoryScreen()));
    // FutureBuilder resolves with fallback list after a frame.
    await t.pump(const Duration(milliseconds: 50));
    expect(find.byKey(const Key('category.pizza')), findsOneWidget);
    expect(find.byKey(const Key('category.pasta')), findsOneWidget);
    expect(find.byKey(const Key('category.burger')), findsOneWidget);
    expect(find.byKey(const Key('category.salads')), findsOneWidget);
    expect(find.byKey(const Key('category.hot')), findsOneWidget);
    expect(find.byKey(const Key('category.cold')), findsOneWidget);
  });
}

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:monti_customer/screens/02_language.dart';

import '../helpers/pump_screen.dart';

void main() {
  testWidgets('Language screen exposes 4 cards and a Next CTA', (t) async {
    await t.pumpWidget(pumpScreen(const LanguageSelectionScreen()));
    expect(find.byKey(const Key('language.card.en')), findsOneWidget);
    expect(find.byKey(const Key('language.card.th')), findsOneWidget);
    expect(find.byKey(const Key('language.card.zh')), findsOneWidget);
    expect(find.byKey(const Key('language.card.ja')), findsOneWidget);
    expect(find.byKey(const Key('language.next')), findsOneWidget);
  });

  testWidgets('Tapping a tile selects it and enables Next', (t) async {
    await t.pumpWidget(pumpScreen(const LanguageSelectionScreen()));
    await t.tap(find.byKey(const Key('language.card.th')));
    await t.pump();
    final btn = t.widget<ElevatedButton>(find.byKey(const Key('language.next')));
    expect(btn.onPressed, isNotNull);
  });
}

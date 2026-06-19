import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:monti_customer/screens/08_thank_you.dart';

import '../helpers/pump_screen.dart';

void main() {
  testWidgets('Thank You renders headline + order number + continue', (t) async {
    await t.pumpWidget(
      pumpScreen(const ThankYouScreen(orderNumber: 'MO-2026-00481')),
    );
    expect(find.text('Thank You'), findsOneWidget);
    expect(find.textContaining('MO-2026-00481'), findsOneWidget);
    expect(find.byKey(const Key('thank.continue')), findsOneWidget);
  });
}

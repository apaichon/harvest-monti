import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:monti_customer/screens/01_welcome.dart';

import '../helpers/pump_screen.dart';

void main() {
  testWidgets('Welcome screen renders headline, prompt, and two CTAs', (t) async {
    await t.pumpWidget(pumpScreen(const WelcomeScreen()));
    expect(find.text("Hi! I'm Monti"), findsOneWidget);
    expect(find.text('What would you like to do?'), findsOneWidget);
    expect(find.byKey(const Key('welcome.start_order')), findsOneWidget);
    expect(find.byKey(const Key('welcome.scan_qr')), findsOneWidget);
    expect(find.textContaining('v1.0'), findsOneWidget);
  });
}

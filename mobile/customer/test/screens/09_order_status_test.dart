import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:monti_customer/models/order.dart';
import 'package:monti_customer/screens/09_order_status.dart';
import 'package:monti_customer/state/order_status_controller.dart';
import 'package:monti_customer/state/providers.dart';

import '../helpers/pump_screen.dart';

void main() {
  testWidgets('Order Status header + ETA + 3-step stepper render', (t) async {
    await t.pumpWidget(
      pumpScreen(const OrderStatusScreen(orderNumber: 'MO-2026-00481')),
    );
    await t.pump();
    expect(find.textContaining('MO-2026-00481'), findsOneWidget);
    expect(find.textContaining('15-20'), findsOneWidget);
    expect(find.text('Order Received'), findsOneWidget);
    expect(find.text('Preparing'), findsOneWidget);
    expect(find.text('Ready'), findsOneWidget);
  });

  testWidgets('WS drop falls back to polling indicator', (t) async {
    final container = ProviderContainer();
    await t.pumpWidget(UncontrolledProviderScope(
      container: container,
      child: pumpScreen(const OrderStatusScreen(orderNumber: 'MO-X')),
    ));
    await t.pump();
    // Simulate WS drop using the controller's test helper.
    container
        .read(orderStatusProvider('MO-X').notifier)
        .simulateWsDrop();
    await t.pump();
    final state = container.read(orderStatusProvider('MO-X'));
    expect(
      state.connection == ConnectionMode.polling ||
          state.connection == ConnectionMode.offline,
      isTrue,
    );
  });

  testWidgets('Pushed status advances the stepper to preparing', (t) async {
    final container = ProviderContainer();
    await t.pumpWidget(UncontrolledProviderScope(
      container: container,
      child: pumpScreen(const OrderStatusScreen(orderNumber: 'MO-Y')),
    ));
    await t.pump();
    container
        .read(orderStatusProvider('MO-Y').notifier)
        .pushStatus(const OrderStatus(
          orderNumber: 'MO-Y',
          currentStep: OrderStep.preparing,
        ));
    await t.pump();
    expect(find.byType(CircularProgressIndicator), findsOneWidget);
  });
}

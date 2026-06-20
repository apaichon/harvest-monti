import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:monti_customer/models/menu.dart';
import 'package:monti_customer/models/order.dart';
import 'package:monti_customer/screens/01_welcome.dart';
import 'package:monti_customer/screens/02_language.dart';
import 'package:monti_customer/screens/03_choose_category.dart';
import 'package:monti_customer/screens/04_menu_list.dart';
import 'package:monti_customer/screens/05_item_detail.dart';
import 'package:monti_customer/screens/06_cart.dart';
import 'package:monti_customer/screens/07_checkout_payment.dart';
import 'package:monti_customer/screens/08_thank_you.dart';
import 'package:monti_customer/screens/09_order_status.dart';
import 'package:monti_customer/state/order_status_controller.dart';
import 'package:monti_customer/state/providers.dart';

import '../helpers/pump_screen.dart';

/// Golden image diffs — one per screen per DES-0008 §3.1-§3.9.
/// Baselines live alongside this file as `screen_3_<n>.png` and are managed
/// via `flutter test --update-goldens` on a clean run.
///
/// Frame size set to 390x844 (iPhone 15 simulator from TEST-0011 preconditions).
void main() {
  setUp(() {
    // Tests force-physical pixel ratio 1.0 so golden output is deterministic.
    TestWidgetsFlutterBinding.ensureInitialized();
  });

  Future<void> _pumpAndGold(
    WidgetTester t,
    Widget child,
    String goldenFile, {
    ProviderContainer? container,
  }) async {
    await t.binding.setSurfaceSize(const Size(390, 844));
    final scope = container == null
        ? child
        : UncontrolledProviderScope(container: container, child: child);
    await t.pumpWidget(pumpScreen(scope));
    await t.pumpAndSettle(const Duration(milliseconds: 100));
    await expectLater(find.byType(MaterialApp), matchesGoldenFile(goldenFile));
  }

  testWidgets('golden 3.1 Welcome', (t) async {
    await _pumpAndGold(t, const WelcomeScreen(), 'screen_3_1.png');
  });

  testWidgets('golden 3.2 Language', (t) async {
    await _pumpAndGold(t, const LanguageSelectionScreen(), 'screen_3_2.png');
  });

  testWidgets('golden 3.3 Categories', (t) async {
    await _pumpAndGold(t, const ChooseCategoryScreen(), 'screen_3_3.png');
  });

  testWidgets('golden 3.4 Menu List', (t) async {
    await _pumpAndGold(
        t, const MenuListScreen(categoryId: 'pizza'), 'screen_3_4.png');
  });

  testWidgets('golden 3.5 Item Detail', (t) async {
    const item = MenuItem(
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
    await _pumpAndGold(t, const ItemDetailScreen(item: item), 'screen_3_5.png');
  });

  testWidgets('golden 3.6 Cart', (t) async {
    final container = ProviderContainer();
    container.read(cartProvider.notifier).addLine(
          item: const MenuItem(
              id: 'latte', name: 'Iced Latte', priceTHB: 120),
          qty: 1,
          unitPriceTHB: 120,
          selectedModifierOptionIds: const {},
        );
    container.read(cartProvider.notifier).addLine(
          item: const MenuItem(
              id: 'margherita', name: 'Margherita Pizza', priceTHB: 240),
          qty: 2,
          unitPriceTHB: 240,
          selectedModifierOptionIds: const {},
        );
    await _pumpAndGold(t, const CartScreen(), 'screen_3_6.png',
        container: container);
  });

  testWidgets('golden 3.7 Checkout Payment', (t) async {
    await _pumpAndGold(
        t, const CheckoutPaymentScreen(), 'screen_3_7.png');
  });

  testWidgets('golden 3.8 Thank You', (t) async {
    await _pumpAndGold(
        t,
        const ThankYouScreen(orderNumber: 'MO-2026-00481'),
        'screen_3_8.png');
  });

  testWidgets('golden 3.9 Order Status', (t) async {
    final container = ProviderContainer();
    container
        .read(orderStatusProvider('MO-2026-00481').notifier)
        .pushStatus(const OrderStatus(
          orderNumber: 'MO-2026-00481',
          currentStep: OrderStep.preparing,
        ));
    await _pumpAndGold(
        t,
        const OrderStatusScreen(orderNumber: 'MO-2026-00481'),
        'screen_3_9.png',
        container: container);
  });
}

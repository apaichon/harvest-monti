import 'package:flutter_test/flutter_test.dart';
import 'package:monti_customer/models/cart.dart';
import 'package:monti_customer/models/menu.dart';
import 'package:monti_customer/state/cart_controller.dart';

void main() {
  test('CartTotals: service 10%, tax 7%, discount subtracts', () {
    final c = CartController();
    c.addLine(
      item: const MenuItem(id: 'a', name: 'A', priceTHB: 100),
      qty: 6,
      unitPriceTHB: 100,
      selectedModifierOptionIds: const {},
    );
    expect(c.state.totals.subtotalTHB, 600);
    expect(c.state.totals.serviceChargeTHB, 60);
    expect(c.state.totals.taxTHB, 42);
    expect(c.state.totals.totalTHB, 702);

    c.applyPromo(const AppliedPromo(code: 'LUNCH20', discountTHB: 50));
    expect(c.state.totals.discountTHB, 50);
    expect(c.state.totals.totalTHB, 652);
  });

  test('Removing last line clears promo and returns to empty', () {
    final c = CartController();
    c.addLine(
      item: const MenuItem(id: 'a', name: 'A', priceTHB: 100),
      qty: 1,
      unitPriceTHB: 100,
      selectedModifierOptionIds: const {},
    );
    final lineId = c.state.lines.single.lineId;
    c.applyPromo(const AppliedPromo(code: 'X', discountTHB: 10));
    c.removeLine(lineId);
    expect(c.state.isEmpty, isTrue);
    expect(c.state.promo, isNull);
  });
}

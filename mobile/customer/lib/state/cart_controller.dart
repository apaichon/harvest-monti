import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/cart.dart';
import '../models/menu.dart';

/// Service charge per DES-0008 §3.6 = 10%.
const double kServiceChargeRate = 0.10;

/// Tax 7% per DES-0008 §3.6.
const double kTaxRate = 0.07;

@immutable
class CartState {
  const CartState({
    this.lines = const [],
    this.promo,
  });

  final List<CartLine> lines;
  final AppliedPromo? promo;

  bool get isEmpty => lines.isEmpty;

  CartTotals get totals {
    final subtotal = lines.fold<int>(0, (s, l) => s + l.lineTotalTHB);
    final service = (subtotal * kServiceChargeRate).round();
    final tax = (subtotal * kTaxRate).round();
    final discount = promo?.discountTHB ?? 0;
    final total = (subtotal + service + tax - discount).clamp(0, 1 << 30);
    return CartTotals(
      subtotalTHB: subtotal,
      serviceChargeTHB: service,
      taxTHB: tax,
      discountTHB: discount,
      totalTHB: total,
    );
  }

  CartState copyWith({List<CartLine>? lines, AppliedPromo? promo, bool clearPromo = false}) =>
      CartState(
        lines: lines ?? this.lines,
        promo: clearPromo ? null : (promo ?? this.promo),
      );
}

/// Cart state machine (DES-0008 §5.1). The empty/adding/checking_out
/// distinction is implicit: `isEmpty == true` is `empty`, otherwise `adding`.
/// Checkout transitions are owned by the routing layer.
class CartController extends StateNotifier<CartState> {
  CartController() : super(const CartState());

  /// Append a line. Caller has already resolved required modifiers.
  void addLine({
    required MenuItem item,
    required int qty,
    required int unitPriceTHB,
    required Set<String> selectedModifierOptionIds,
  }) {
    final lineId = 'L${DateTime.now().microsecondsSinceEpoch}';
    final line = CartLine(
      lineId: lineId,
      item: item,
      qty: qty,
      unitPriceTHB: unitPriceTHB,
      selectedModifierOptionIds: selectedModifierOptionIds,
    );
    state = state.copyWith(lines: [...state.lines, line]);
  }

  /// Update qty inline (cart screen stepper). Removing to 0 removes the line.
  void updateQty(String lineId, int qty) {
    if (qty <= 0) {
      removeLine(lineId);
      return;
    }
    state = state.copyWith(
      lines: [
        for (final l in state.lines)
          if (l.lineId == lineId) l.copyWith(qty: qty) else l,
      ],
    );
  }

  void removeLine(String lineId) {
    final next = state.lines.where((l) => l.lineId != lineId).toList();
    state = next.isEmpty
        ? state.copyWith(lines: next, clearPromo: true)
        : state.copyWith(lines: next);
  }

  /// Apply a promo locally (server-side reconciliation is final per REQ-0012).
  void applyPromo(AppliedPromo promo) {
    state = state.copyWith(promo: promo);
  }

  void clearPromo() {
    state = state.copyWith(clearPromo: true);
  }

  void reset() {
    state = const CartState();
  }
}

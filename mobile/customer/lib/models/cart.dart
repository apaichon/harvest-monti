import 'package:flutter/foundation.dart';

import 'menu.dart';

/// One line in the cart. Modifier deltas are pre-baked into [unitPriceTHB].
@immutable
class CartLine {
  const CartLine({
    required this.lineId,
    required this.item,
    required this.qty,
    required this.unitPriceTHB,
    this.selectedModifierOptionIds = const {},
  });

  final String lineId;
  final MenuItem item;
  final int qty;
  final int unitPriceTHB;
  final Set<String> selectedModifierOptionIds;

  int get lineTotalTHB => unitPriceTHB * qty;

  CartLine copyWith({int? qty}) => CartLine(
        lineId: lineId,
        item: item,
        qty: qty ?? this.qty,
        unitPriceTHB: unitPriceTHB,
        selectedModifierOptionIds: selectedModifierOptionIds,
      );
}

/// Aggregate totals — DES-0008 §3.6 / §3.7 line breakdown.
@immutable
class CartTotals {
  const CartTotals({
    required this.subtotalTHB,
    required this.serviceChargeTHB,
    required this.taxTHB,
    required this.discountTHB,
    required this.totalTHB,
  });

  final int subtotalTHB;
  final int serviceChargeTHB;
  final int taxTHB;
  final int discountTHB;
  final int totalTHB;

  static const empty = CartTotals(
    subtotalTHB: 0,
    serviceChargeTHB: 0,
    taxTHB: 0,
    discountTHB: 0,
    totalTHB: 0,
  );

  bool get hasDiscount => discountTHB > 0;
}

/// Sticky promo applied to the cart.
@immutable
class AppliedPromo {
  const AppliedPromo({required this.code, required this.discountTHB});

  final String code;
  final int discountTHB;
}

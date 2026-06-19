import 'package:flutter/foundation.dart';

/// Payment method tile id (DES-0008 §3.7).
enum PaymentMethod { creditCard, applePay, wallet, cash }

extension PaymentMethodId on PaymentMethod {
  String get apiValue => switch (this) {
        PaymentMethod.creditCard => 'credit_card',
        PaymentMethod.applePay => 'apple_pay',
        PaymentMethod.wallet => 'wallet',
        PaymentMethod.cash => 'cash',
      };
}

/// Order status step (DES-0008 §3.9).
enum OrderStep { received, preparing, ready }

@immutable
class OrderStatus {
  const OrderStatus({
    required this.orderNumber,
    required this.currentStep,
    this.etaMinLow = 15,
    this.etaMinHigh = 20,
    this.receivedAtLocal,
  });

  final String orderNumber;
  final OrderStep currentStep;
  final int etaMinLow;
  final int etaMinHigh;
  final String? receivedAtLocal;

  OrderStatus copyWith({OrderStep? currentStep}) => OrderStatus(
        orderNumber: orderNumber,
        currentStep: currentStep ?? this.currentStep,
        etaMinLow: etaMinLow,
        etaMinHigh: etaMinHigh,
        receivedAtLocal: receivedAtLocal,
      );

  factory OrderStatus.fromJson(Map<String, dynamic> json) => OrderStatus(
        orderNumber: json['order_number'] as String,
        currentStep: _parseStep(json['status'] as String?),
        etaMinLow: (json['eta_min_low'] ?? 15) as int,
        etaMinHigh: (json['eta_min_high'] ?? 20) as int,
        receivedAtLocal: json['received_at_local'] as String?,
      );

  static OrderStep _parseStep(String? raw) => switch (raw) {
        'preparing' => OrderStep.preparing,
        'ready' => OrderStep.ready,
        _ => OrderStep.received,
      };
}
